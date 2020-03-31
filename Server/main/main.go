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
	l := flag.Int("lag", 0, "integer for how much lag the server should emulate")

	rr := flag.Int("refresh", 33, "integer for how many times the server should broadcast messages per second")

	flag.Parse()

	lag = *l
	refreshRate = *rr

	r := gin.Default()
	m = melody.New()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "alive",
		})
	})

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		var f interface{}
		err := json.Unmarshal(msg, &f)
		if err != nil {
			panic(err)
		}

		u := f.(map[string]interface{})

		u["duration"] = refreshRate
		msg, err = json.Marshal(u)

		if (u["command"] == "NEW_PLAYER") || (u["command"] == "REFRESH_PLAYER") {
			go func() {
				time.Sleep(time.Duration(lag) * time.Millisecond)
				m.Broadcast(msg)
			}()
			return
		}
		id := u["player_id"].(float64)
		if err != nil {
			panic(err)
		}
		mutex.Lock()
		playerMap[id] = msg
		mutex.Unlock()
	})

	go sendLoop()
	fmt.Printf("Current settings: \n lag: %d \n refresh rate: %d \n",lag,refreshRate)
	r.Run(":5000")
}

func sendLoop() {
	time.Sleep(time.Duration(refreshRate) * time.Millisecond)
	mutex.Lock()
	for i, msg := range playerMap {
		go func() {
			time.Sleep(time.Duration(lag) * time.Millisecond)
			m.Broadcast(msg)
		}()
		delete(playerMap,i)
	}
	mutex.Unlock()
	go sendLoop()
}
