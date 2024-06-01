package core

import (
	"context"

	"github.com/google/uuid"
)

type Action = string

const (
	ActionFinish  = "finish"
	ActionRespond = "respond"
	ActionListen  = "listen"
)

type Client interface {
	AcquireSession(ctx context.Context, agent, gateway string) (uuid.UUID, error)
	ReleaseSession(ctx context.Context, sessionID uuid.UUID) error
	Process(ctx context.Context, sessionID uuid.UUID) (Handler, error)
}

type ProcessInputHandler interface {
	Handle(input *ProcessInput) error
}

type ProcessOutputHandler interface {
	HasMessage() bool
	Await() (*ProcessOutput, error)
}

type Handler interface {
	GetInputHandler() ProcessInputHandler
	GetOutputHandler() ProcessOutputHandler
	Error() error
	Close()
	Done() <-chan struct{}
}

type ProcessInput struct {
	UserInput string
	RequestID uuid.UUID
}

type ProcessOutput struct {
	Action       Action
	UserResponse string
}
