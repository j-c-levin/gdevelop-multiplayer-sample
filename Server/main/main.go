package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
	"sync"
	"time"
)

var lag int
var refreshRate int
var m *melody.Melody
var playerMap = make(map[float64][]byte)
var mutex sync.Mutex

func main() {
	// Setting flags and providing defaults
	l := flag.Int("lag", 0, "integer for how much lag the server should emulate")
	rr := flag.Int("refresh", 5, "integer for number of milliseconds between message broadcasts")

	flag.Parse()

	// Setting the variables
	lag = *l
	refreshRate = *rr

	// Gin is the server, melody is for websockets
	r := gin.Default()
	m = melody.New()

	// Healthcheck endpoint for testing
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "alive",
		})
	})

	// Go to this path to connect to a websocket
	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	// Run every time a client sends a websocket message
	m.HandleMessage(func(s *melody.Session, messageBytes []byte) {
		// Extracts the JSON message into something Go can interact with
		var f interface{}
		err := json.Unmarshal(messageBytes, &f)
		if err != nil {
			panic(err)
		}

		message := f.(map[string]interface{})

		// We manually set the time difference between one message and the next
		message["duration"] = refreshRate
		messageBytes, err = json.Marshal(message)

		// We don't want to buffer new player or refresh commands
		if (message["command"] == "NEW_PLAYER") || (message["command"] == "REFRESH_PLAYER") {
			go func() {
				// Mimic lag, sleep the code for 'lag' milliseconds
				time.Sleep(time.Duration(lag) * time.Millisecond)
				m.Broadcast(messageBytes)
			}()
			return
		}

		id := message["player_id"].(float64)
		if err != nil {
			panic(err)
		}
		mutex.Lock()
		// Store the message into a map so we can send it later
		playerMap[id] = messageBytes
		mutex.Unlock()
	})

	go sendLoop()
	fmt.Printf("Current settings: \n lag: %d \n refresh rate: %d \n",lag,refreshRate)
	r.Run(":5000")
}

func sendLoop() {
	// Mimic server refresh rate by sleeping for 'refresh' milliseconds
	time.Sleep(time.Duration(refreshRate) * time.Millisecond)

	mutex.Lock()
	for i, msg := range playerMap {
		go func() {
			// Mimic lag, sleep the code for 'lag' milliseconds
			time.Sleep(time.Duration(lag) * time.Millisecond)
			m.Broadcast(msg)
		}()
		delete(playerMap,i)
	}
	mutex.Unlock()
	go sendLoop()
}
