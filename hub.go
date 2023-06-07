package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/juancortelezzi/websockets/coords"
	"github.com/juancortelezzi/websockets/redis"
	goredis "github.com/redis/go-redis/v9"

	"github.com/uber/h3-go/v4"
)

var ctx = context.Background()

type ConnectionHub struct {
	Clients    map[int32]*Client
	FromSocket chan *MsgFromClient
	FromServer chan *MsgFromServer
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *ConnectionHub {
	return &ConnectionHub{
		Clients:    make(map[int32]*Client),
		FromSocket: make(chan *MsgFromClient),
		FromServer: make(chan *MsgFromServer),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func DeleteClient(clients map[int32]*Client, client *Client) {
	close(client.Send)
	delete(clients, client.ID)
}

func (h *ConnectionHub) Run() {
	for {
		select {
		case msg := <-h.FromSocket:
			switch msg.Type {
			case TypeLocation:

				data, err := parseMessage[LocationData](msg.Data)
				if err != nil {
					log.Printf("parse error: %v", err)
					return
				}

				go h.toTakeOrNotToTakeAPhotoThatIsTheQuestion(msg.Client, data)
				go h.locationSaver(msg.Client, data)

			default:
				if client, exists := h.Clients[msg.Client.ID]; exists {
					DeleteClient(h.Clients, client)
				}
			}

		case msg := <-h.FromServer:
			if client, exists := h.Clients[msg.To]; exists {
				client.Send <- msg.toBytes()
				msg.OneShot <- nil
				close(msg.OneShot)
				continue
			}
			msg.OneShot <- errors.New("Client not found")
			close(msg.OneShot)

		case client := <-h.Register:
			h.Clients[client.ID] = client
			log.Printf("registered client `%d` len: %d\n", client.ID, len(h.Clients))

		case client := <-h.Unregister:
			if _, exists := h.Clients[client.ID]; exists {
				DeleteClient(h.Clients, client)
			}
			log.Printf("removed client `%d` len: %d\n", client.ID, len(h.Clients))
		}
	}
}

func (h *ConnectionHub) toTakeOrNotToTakeAPhotoThatIsTheQuestion(
	client *Client,
	data *LocationData,
) {
	dbg_now := time.Now()

	position := h3.NewLatLng(data.Lat, data.Lng).Cell(coords.RESOLUTION)

	// bus_stops:<org_id>
	bus_stops_key := fmt.Sprintf("bus_stops:%d", client.ID)

	{
		// Does redis have the bus stops?
		numOfKeys, err := redis.Redis.Exists(ctx, bus_stops_key).Result()
		if err != nil {
			log.Panicln("crash and burn, no redis!", err)
		}

		if numOfKeys < 1 {
			// There are no bus stops in redis, hydrate them.

			// This should be done before the pipelineing of redis to avoid blocking.
			coordsFromSqeel := coords.GetMeArrayBabyyy()

			pipe := redis.Redis.Pipeline()

			pipe.SAdd(ctx, bus_stops_key, coordsFromSqeel)

			appliedCmd := pipe.Expire(ctx, bus_stops_key, time.Minute*5)

			_, err := pipe.Exec(ctx)

			// .Val can only be used after the Exec
			if err != nil || !appliedCmd.Val() {
				log.Panicln("crash and burn, no redis!", err)
			}
		}
	}

	disk := position.GridDisk(1)
	disk_strs := make([]string, len(disk))
	for _, hex := range disk {
		disk_strs = append(disk_strs, hex.String())
	}

	// if any hex in grid contains a bus stop, take a photo
	bools, err := redis.Redis.SMIsMember(ctx, bus_stops_key, disk_strs).Result()
	if err != nil {
		log.Panicln("crash and burn, no redis!", err)
	}

	log.Printf("disk(%d): %v\n", data.ID, disk_strs)
	log.Printf("bools(%d): %v", data.ID, bools)

	for _, b := range bools {
		if b {
			log.Printf("PHOTO(%d): for request<%v>", client.ID, data.ID)
			// client.Send <- NewEventMsg(TypeLocation).toBytes()
			client.Send <- Msg[IDDdata]{
				Type: TypeTakePhoto,
				Data: IDDdata{
					Id: data.ID,
				},
			}.toBytes()
			break
		}
	}

	log.Printf(
		"serviced(%v) cell<%v> in<%#v>\n",
		data.ID,
		position.String(),
		time.Since(dbg_now).String(),
	)
}

func (h *ConnectionHub) locationSaver(client *Client, data *LocationData) {
	position := h3.NewLatLng(data.Lat, data.Lng).Cell(coords.RESOLUTION)

	// bus_last_position:<device_id> = "timestamp::h3.cellID"
	last_position_key := fmt.Sprintf(
		"bus_last_position:%d",
		client.ID,
	)

	last_position_str, err := redis.Redis.Get(ctx, last_position_key).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {

			log.Printf(
				"LOCATION(%d): cell<%v> empty redis\n",
				data.ID,
				position,
			)
			savePositionToSqeelAndUpdateRedisCache(last_position_key, position)
			return
		}
		log.Panicln("crash and burn, no redis!:", err)
	}

	last_position_kv := strings.Split(last_position_str, "::")
	if len(last_position_kv) != 2 {
		panic("I stored the wrong thing in redis, crash and burn!")
	}

	last_position_timestamp, err := time.Parse(time.RFC3339, last_position_kv[0])
	if err != nil {
		panic("I stored the wrong timestamp in redis, crash and burn!")
	}

	last_position_cell := h3.Cell(int64(h3.IndexFromString(last_position_kv[1])))
	if !last_position_cell.IsValid() {
		panic("I stored the wrong cell in redis, crash and burn!")
	}

	timed_out := last_position_timestamp.Add(time.Minute * 5).
		Before(time.Now())

	changed_position := position != last_position_cell

	log.Printf(
		"LOCATION(%d): cell<%v> changed_position<%v> | timed_out<%v>\n",
		data.ID,
		position,
		changed_position,
		timed_out,
	)

	if timed_out || changed_position {
		// save position to sqeel and update redis cache
		savePositionToSqeelAndUpdateRedisCache(last_position_key, position)
	}
}

func savePositionToSqeelAndUpdateRedisCache(
	last_position_key string,
	position h3.Cell,
) {
	timestamp_string := time.Now().Format(time.RFC3339)
	cell_string := position.String()
	value := timestamp_string + "::" + cell_string
	if err := redis.Redis.Set(ctx, last_position_key, value, 0).Err(); err != nil {
		log.Panicln("crash and burn, no redis!", err)
	}
}
