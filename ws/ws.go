package ws

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type state_Data struct {
	State         int `json:"state"`
	Volume        int `json:"volume"`
	Repeat        int `json:"repeat"`
	Single        int `json:"single"`
	Consume       int `json:"consume"`
	Random        int `json:"random"`
	SongPos       int `json:"songpos"`
	ElapsedTime   int `json:"elapsedTime"`
	TotalTime     int `json:"totalTime"`
	CurrentSongID int `json:"currentsongid"`
}

type song_Data struct {
	Position int    `json:"pos"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
}
type queue_Data struct {
	ID        int    `json:"id"`
	Position  int    `json:"pos"`
	Durration int    `json:"duration"`
	Title     string `json:"title"`
}

// Set the websocket address.
var Url string = "ws://127.0.0.1"

// Declare Public Variables we can use to get the current data from other classes eg. dispremote/handler.go
var WSState = state_Data{}
var WSQueue = []queue_Data{}
var WSSongInfo song_Data

func setWSSongInfo(songInfo song_Data) {
	WSSongInfo = songInfo
}
func setWSState(stateData state_Data) {
	WSState = stateData
}
func setWSQueue(queueData []queue_Data) {
	WSQueue = queueData
}

func Start() {
	//Creates Reading Websocket Handler and parses message based on type
	conn, _, err := websocket.DefaultDialer.Dial(Url, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer conn.Close()
	for {
		// Read message from WebSocket
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// Parse message into the Message struct
		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("Error unmarshaling message:", err)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "song_change":
			var songinfo song_Data
			if err := json.Unmarshal(msg.Data, &songinfo); err != nil {
				log.Println("Error decoding song change:", err)
				continue
			}
			setWSSongInfo(songinfo)

		case "state":
			var statedata state_Data
			if err := json.Unmarshal(msg.Data, &statedata); err != nil {
				log.Println("Error decoding state:", err)
				continue
			}
			setWSState(statedata)

		case "queue":
			var queuenow []queue_Data
			if err := json.Unmarshal(msg.Data, &queuenow); err != nil {
				log.Println("Error decoding queue:", err)
				continue
			}
			setWSQueue(queuenow)

		case "update_queue":
			//ympd lets us know the queue was updated but doesn't send the updated queue to us.
			//Write a message to get the uodated data in a "queue" message
			err := conn.WriteMessage(websocket.TextMessage, []byte("MPD_API_GET_QUEUE,0"))
			if err != nil {
				log.Println("Error sending update_queue message:", err)
			}

		default:
			log.Warn("Unsupported WS message type:", msg.Type)
		}
	}
}

func CommandWS(comm string) {
	var conn = new(websocket.Conn)
	conn, _, err := websocket.DefaultDialer.Dial(Url, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer conn.Close()
	for {
		// Write message to WebSocket
		conn.WriteMessage(websocket.TextMessage, []byte(comm))
		break
	}
}

func NextSong() {
	CommandWS("MPD_API_SET_NEXT")
}
func PrevSong() {
	CommandWS("MPD_API_SET_PREV")
}

func SetPlayingTrack(c int32) {
	var a = "MPD_API_PLAY_TRACK"
	var b = c + 1
	var d = strings.Join([]string{a, strconv.Itoa(int(b))}, ",")
	CommandWS(d)
}
