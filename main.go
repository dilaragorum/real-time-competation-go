package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"sync"
	"time"
)

const (
	CompetationNotStartedState = "NOT_STARTED"
	CompetationStartedState    = "STARTED"
	CompetationFinish          = "FINISH"
)

type Question struct {
	ID              int      `json:"id"`
	Options         []string `json:"options"`
	RightAnswer     string   `json:"right_answer"`
	QuestionContent string   `json:"question_content"`
	IsDisabled      bool     `json:"is_disabled"`
}

type Client struct {
	wsConn     *websocket.Conn
	totalScore int
}

var CompetationState string

var (
	Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	Clients sync.Map //To support concurrent write and read operations

	Questions = []Question{
		{
			ID: 1,
			Options: []string{
				"A", "B", "C", "D", "E",
			},
			RightAnswer:     "C",
			QuestionContent: "Right Answer is C",
		},
		{
			ID: 2,
			Options: []string{
				"A", "B", "C", "D", "E",
			},
			RightAnswer:     "C",
			QuestionContent: "Right Answer is B",
		},
	}
)

func main() {
	CompetationState = CompetationNotStartedState

	e := echo.New()
	e.Use(middleware.RequestID())

	e.GET("/ws", ws)
	e.File("/", "index.html")

	go func() {
		for {
			if CompetationState == CompetationNotStartedState {
				time.Sleep(3 * time.Second)
				var numberOfClients int
				Clients.Range(func(key, value any) bool {
					numberOfClients++
					return true
				})
				var msg string
				if numberOfClients == 1 {
					msg = "waiting for another user"
				} else if numberOfClients == 2 {
					msg = "competition is starting within 3 seconds"
				}

				Clients.Range(func(key, value any) bool {
					client := value.(Client)
					client.wsConn.WriteMessage(websocket.TextMessage, []byte(msg))
					return true
				})

				if numberOfClients == 2 {
					time.Sleep(3 * time.Second)
					CompetationState = CompetationStartedState
				}
			} else if CompetationState == CompetationStartedState {
				for _, q := range Questions {
					q.IsDisabled = false
					Clients.Range(func(key, value any) bool {
						client := value.(Client)
						client.wsConn.WriteJSON(q)
						return true
					})
					time.Sleep(10 * time.Second)
					q.IsDisabled = true
				}

				CompetationState = CompetationFinish
			} else if CompetationState == CompetationFinish {
				leaderBoard := LeaderBoard{}
				Clients.Range(func(key, value any) bool {
					client := value.(Client)
					competitionResult := CompetitionResult{
						RequestID:  fmt.Sprintf("%s", key),
						TotalScore: client.totalScore,
					}
					leaderBoard.CompetitionResults = append(leaderBoard.CompetitionResults, competitionResult)
					return true
				})

				//After calculation sending leaderBoard to All Clients
				Clients.Range(func(key, value any) bool {
					client := value.(Client)
					client.wsConn.WriteJSON(leaderBoard)
					return true
				})
			}
		}
	}()

	e.Logger.Fatal(e.Start(":8080"))
}

type FEMessage struct {
	Id     int    `json:"id"`
	Answer string `json:"answer"`
}

func ws(c echo.Context) error {
	//upgrading to Websocket
	wsConn, err := Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer wsConn.Close()

	//mapping users with their request id
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	Clients.Store(requestID, Client{
		wsConn:     wsConn,
		totalScore: 0,
	})

	//Reading message/ Listening client
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			Clients.Delete(requestID)
			c.Logger().Errorf("Client disconnect msg=%s err=%s", string(message), err.Error())
			return nil
		}

		var FEMes FEMessage
		json.Unmarshal(message, &FEMes)

		for _, val := range Questions {
			if val.ID == FEMes.Id && val.IsDisabled == false {
				load, _ := Clients.Load(requestID)
				client := load.(Client)
				if FEMes.Answer == val.RightAnswer {
					client.totalScore += 10
					Clients.Store(requestID, client)
					fmt.Printf("Right Answer!!! RequestId: %s, TotalScore: %d\n", requestID, client.totalScore)
				} else {
					fmt.Printf("Wrong Answer!! Your Answer is : %s, Right Answer is : %s, RequestId: %s, TotalScore: %d\n", FEMes.Answer, val.RightAnswer, requestID, client.totalScore)
				}
			} else {
				fmt.Printf("Time is out\n")
			}
		}
	}
}

type LeaderBoard struct {
	CompetitionResults []CompetitionResult
}

type CompetitionResult struct {
	RequestID  string `json:"request_id"`
	TotalScore int    `json:"total_score"`
}
