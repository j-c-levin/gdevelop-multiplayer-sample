package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
	"time"
)

const lag = 30
const dropRate = 5
var currentDrop = 0

func main() {
	r := gin.Default()
	m := melody.New()

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
		if (u["command"] != "NEW_PLAYER") && (u["command"] != "REFRESH_PLAYER") {
			currentDrop++
			if currentDrop % dropRate != 0 {
				return
			}
		}

		go func() {
			time.Sleep(lag * time.Millisecond)
			m.Broadcast(msg)
		}()
	})

	r.Run(":5000")
}
