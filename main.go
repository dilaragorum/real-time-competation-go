package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"os"
	"sync"
	"time"
)

var (
	CompetitionState string // "NOT_STARTED", "STARTED", "FINISH"
	Clients          sync.Map
	Questions        []Question
)

func main() {
	e := echo.New()

	e.GET("/ws", ws)
	e.File("/", "./static/index.html")

	go RunCompetition()

	e.Logger.Fatal(e.Start(":8080"))
}

func PrepareQuestions() {
	questionsBytes, err := os.ReadFile("./assets/questions.json")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(questionsBytes, &Questions); err != nil {
		panic(err)
	}
}

func RunCompetition() {
	CompetitionState = CompetitionNotStartedState

	for {
		if CompetitionState == CompetitionNotStartedState {
			time.Sleep(CompetitionStateDuration)

			numberOfClients := CountClient()

			msg := DetermineCompetitionState(numberOfClients)
			BroadcastMessage([]byte(msg))

			if numberOfClients == 2 {
				time.Sleep(CompetitionStartDuration)
				CompetitionState = CompetitionStartedState
			}
		} else if CompetitionState == CompetitionStartedState {
			PrepareQuestions()
			StartSendingQuestions()
			CompetitionState = CompetitionFinish
		} else if CompetitionState == CompetitionFinish {
			leaderBoard := CreateLeaderBoard()
			jsonBytes, _ := json.Marshal(leaderBoard)
			BroadcastMessage(jsonBytes)

			BroadcastMessage([]byte(CompetitionFinishedStateMessage))

			break
		}
	}
}

func DetermineCompetitionState(numberOfClients int) string {
	var message string
	if numberOfClients == 1 {
		message = CompetitionWaitingStateMessage
	} else if numberOfClients == 2 {
		message = CompetitionStartingStateMessage
	}
	return message
}

func CountClient() int {
	var numberOfClients int

	Clients.Range(func(key, value any) bool {
		numberOfClients++
		return true
	})

	return numberOfClients
}

func BroadcastMessage(message []byte) {
	Clients.Range(func(key, value any) bool {
		client := value.(Client)
		client.wsConn.WriteMessage(websocket.TextMessage, message)
		return true
	})
}

func StartSendingQuestions() {
	for i := range Questions {
		Questions[i].IsTimeout = false

		questionDTO := Questions[i].ToDTO()
		questionDTOBytes, _ := json.Marshal(questionDTO)
		BroadcastMessage(questionDTOBytes)

		time.Sleep(QuestionResponseIntervalDuration)

		Questions[i].IsTimeout = true
	}
}

func CreateLeaderBoard() LeaderBoard {
	var leaderBoard LeaderBoard

	Clients.Range(func(key, value any) bool {
		client := value.(Client)

		leaderBoard.CompetitionResults = append(leaderBoard.CompetitionResults, CompetitionResult{
			SessionID:  fmt.Sprintf("%s", key),
			TotalScore: client.totalScore,
		})

		return true
	})

	return leaderBoard
}
