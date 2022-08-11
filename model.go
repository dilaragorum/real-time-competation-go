package main

import (
	"github.com/gorilla/websocket"
)

type Question struct {
	ID            int    `json:"id"`
	CorrectAnswer string `json:"correct_answer"`
	Content       string `json:"question_content"`
	ContentA      string `json:"content_a"`
	ContentB      string `json:"content_b"`
	ContentC      string `json:"content_c"`
	ContentD      string `json:"content_d"`
	IsTimeout     bool   `json:"-"`
}

func (q *Question) ToDTO() QuestionDTO {
	return QuestionDTO{
		ID:       q.ID,
		Content:  q.Content,
		ContentA: q.ContentA,
		ContentB: q.ContentB,
		ContentC: q.ContentC,
		ContentD: q.ContentD,
	}
}

type QuestionDTO struct {
	ID       int    `json:"id"`
	Content  string `json:"question_content"`
	ContentA string `json:"content_a"`
	ContentB string `json:"content_b"`
	ContentC string `json:"content_c"`
	ContentD string `json:"content_d"`
}

type Client struct {
	wsConn     *websocket.Conn
	totalScore int
}

type ClientMessage struct {
	QuestionId int    `json:"id"`
	Answer     string `json:"answer"`
}

type LeaderBoard struct {
	CompetitionResults []CompetitionResult
}

type CompetitionResult struct {
	SessionID  string `json:"session_id"`
	TotalScore int    `json:"total_score"`
}
