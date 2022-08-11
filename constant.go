package main

import "time"

const (
	CompetitionNotStartedState = "NOT_STARTED"
	CompetitionStartedState    = "STARTED"
	CompetitionFinish          = "FINISH"

	CompetitionFinishedStateMessage = "competition is over"
	CompetitionWaitingStateMessage  = "waiting for another user"
	CompetitionStartingStateMessage = "competition is starting within 3 seconds"

	CompetitionStateDuration         = 3 * time.Second
	CompetitionStartDuration         = 3 * time.Second
	QuestionResponseIntervalDuration = 10 * time.Second
	ScoreForCorrectAnswer            = 10
)
