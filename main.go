package main

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

type UserIdentifier struct {
	id atomic.Int32
}

func (u *UserIdentifier) GetId() int32 {
	return u.id.Add(1)
}

var userIdentifier UserIdentifier

func main() {
	hub := NewHub()
	go hub.Run()

	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.

		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/takephoto/:id", func(c *fiber.Ctx) error {
		log.Println("hit")
		rawid := c.Params("id")
		id, err := strconv.Atoi(rawid)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		oneShot := make(chan error)
		hub.FromServer <- &MsgFromServer{
			To:      int32(id), // HACK: this may fail.
			OneShot: oneShot,
			Msg:     NewEventMsg(TypeTakePhoto),
		}

		select {
		case err := <-oneShot:
			if err != nil {
				return fiber.NewError(http.StatusBadRequest, err.Error())
			}
			return c.SendStatus(http.StatusOK)
		case <-time.After(1 * time.Second):
			return fiber.NewError(http.StatusInternalServerError, "timeout")
		}
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		client := &Client{
			ID:   userIdentifier.GetId(),
			Hub:  hub,
			Conn: c,
			Send: make(chan []byte, 256),
		}

		client.Hub.Register <- client

		go client.WritePump()
		client.ReadPump()
	}, websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}))

	log.Fatal(app.Listen(":4000"))
}
