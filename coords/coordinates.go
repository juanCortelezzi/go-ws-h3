package coords

import "github.com/uber/h3-go/v4"

const RESOLUTION = 10

type Coord struct {
	Recorrido string  `json:"recorrido"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Cell      h3.Cell `json:"-"`
}

func newCoord(recorrdio string, lat float64, lng float64) Coord {
	return Coord{
		Recorrido: recorrdio,
		Lat:       lat,
		Lng:       lng,
		Cell:      h3.NewLatLng(lat, lng).Cell(RESOLUTION),
	}
}

func GetMeHashmapBabyyyy() map[h3.Cell]Coord {
	hashmap := make(map[h3.Cell]Coord, 0)
	for _, coord := range CoordinatesLineaD {
		hashmap[coord.Cell] = coord
	}
	for _, coord := range CoordinatesLineaB {
		hashmap[coord.Cell] = coord
	}
	for _, coord := range CoordinatesLineaC {
		hashmap[coord.Cell] = coord
	}
	return hashmap
}

// This would be replaced by a SQEEL call
func GetMeArrayBabyyy() []string {
	arr := make([]string, 0)
	for _, coord := range CoordinatesLineaD {
		arr = append(arr, coord.Cell.String())
	}
	// for _, coord := range CoordinatesLineaB {
	// 	arr = append(arr, coord.Cell.String())
	// }
	// for _, coord := range CoordinatesLineaC {
	// 	arr = append(arr, coord.Cell.String())
	// }
	return arr
}

var CoordinatesLineaD = [...]Coord{
	// lat and lon are flipped (flipo)
	newCoord("D", -34.55584270788265, -58.46218465983149),
	newCoord("D", -34.56244610423889, -58.45637430014038),
	newCoord("D", -34.56630614206709, -58.45199716393931),
	newCoord("D", -34.570279928424924, -58.44434436457492),
	newCoord("D", -34.57558167809432, -58.43467054067577),
	newCoord("D", -34.57850456981334, -58.42559696101462),
	newCoord("D", -34.58148604012713, -58.421098048873134),
	newCoord("D", -34.58516120713579, -58.41595971260905),
	newCoord("D", -34.58829242028046, -58.41122019217053),
	newCoord("D", -34.59174497634508, -58.40706844206811),
	newCoord("D", -34.59443353640935, -58.40253797341393),
	newCoord("D", -34.599776767874204, -58.397753965623195),
	newCoord("D", -34.599647788121985, -58.39267111581988),
	newCoord("D", -34.60179641599479, -58.384794769111835),
	newCoord("D", -34.60444061032262, -58.380266667740855),
	newCoord("D", -34.60761882590082, -58.37426363487336),
}

var CoordinatesLineaC = [...]Coord{
	newCoord("C", -34.591176692941595, -58.374790187097574),
	newCoord("C", -34.595061547827605, -58.37787536340444),
	newCoord("C", -34.60227003174365, -58.378124417919096),
	newCoord("C", -34.60477828846878, -58.37949785944903),
	newCoord("C", -34.60897308300395, -58.380628089856316),
	newCoord("C", -34.61249240821683, -58.3805010273409),
	newCoord("C", -34.618074116281946, -58.38012219550376),
	newCoord("C", -34.622107623995596, -58.37992935170675),
	newCoord("C", -34.62738185484165, -58.381116812732074),
}

var CoordinatesLineaB = [...]Coord{
	newCoord("B", -34.57399315122806, -58.48680772610702),
	newCoord("B", -34.57798801938457, -58.48075318364164),
	newCoord("B", -34.581189808218085, -58.47433869018563),
	newCoord("B", -34.58411147744602, -58.46638020359141),
	newCoord("B", -34.586739658439946, -58.455482154042926),
	newCoord("B", -34.59174348470706, -58.447491577888),
	newCoord("B", -34.599072109193514, -58.43969110837304),
	newCoord("B", -34.60214766246166, -58.43139623952922),
	newCoord("B", -34.60319673206347, -58.4208616983951),
	newCoord("B", -34.604082987036456, -58.411351009581324),
	newCoord("B", -34.60456888097357, -58.405393069600436),
	newCoord("B", -34.60463756505772, -58.399443213611704),
	newCoord("B", -34.604454453880926, -58.39290266516748),
	newCoord("B", -34.60403330065313, -58.38707403546648),
	newCoord("B", -34.60370369762329, -58.38116215907441),
	newCoord("B", -34.603374100319094, -58.37456502745073),
	newCoord("B", -34.602999728691444, -58.369633617129736),
}
