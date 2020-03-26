package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
	"sync"
	"time"
)

const lag = 0
const refreshRate = 500
var m *melody.Melody
var playerMap = make(map[float64][]byte)
var mutex sync.Mutex

func main() {
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
		if (u["command"] == "NEW_PLAYER") || (u["command"] == "REFRESH_PLAYER") {
			go func() {
				time.Sleep(lag * time.Millisecond)
				m.Broadcast(msg)
			}()
			return
		}
		id := u["player_id"].(float64)
		mutex.Lock()
		playerMap[id] = msg
		mutex.Unlock()
	})

	go sendLoop()
	r.Run(":5000")
}

func sendLoop() {
	time.Sleep(refreshRate * time.Millisecond)
	mutex.Lock()
	for _, msg := range playerMap {
		go func() {
			time.Sleep(lag * time.Millisecond)
			m.Broadcast(msg)
		}()
	}
	mutex.Unlock()
	go sendLoop()
}
