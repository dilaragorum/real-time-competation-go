package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ws(c echo.Context) error {
	numberOfClients := CountClient()
	if numberOfClients >= 2 {
		return c.String(http.StatusBadRequest, "")
	}

	wsConn, err := Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer wsConn.Close()

	sessionID := IDGenerator()

	Clients.Store(sessionID, Client{
		wsConn:     wsConn,
		totalScore: 0,
	})

	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			Clients.Delete(sessionID)
			c.Logger().Errorf("Client disconnect msg=%s err=%s", string(message), err.Error())
			return nil
		}

		HandleClientAnswer(sessionID, message)
	}
}

func HandleClientAnswer(sessionID string, message []byte) {
	var ClientMsg ClientMessage
	json.Unmarshal(message, &ClientMsg)

	for _, question := range Questions {
		if question.ID == ClientMsg.QuestionId {
			if question.IsTimeout == true {
				fmt.Println("Response Time is out")
			} else {
				load, _ := Clients.Load(sessionID)
				client := load.(Client)
				if ClientMsg.Answer == question.CorrectAnswer {
					client.totalScore += ScoreForCorrectAnswer
					Clients.Store(sessionID, client)
					fmt.Printf("Right Answer!!! SessionId: %s, TotalScore: %d\n", sessionID, client.totalScore)
				} else {
					fmt.Printf("Wrong Answer!! Your Answer is : %s, Right Answer is : %s, SessionId: %s, TotalScore: %d\n", ClientMsg.Answer, question.CorrectAnswer, sessionID, client.totalScore)
				}
			}
		}
	}
}
