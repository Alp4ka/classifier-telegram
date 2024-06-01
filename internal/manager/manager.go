package manager

import (
	"context"
	"fmt"

	"github.com/Alp4ka/classifier-telegram/internal/interactions/core"
	"github.com/google/uuid"
)

type Manager interface {
	core.Client
}

type managerImpl struct {
	core.Client
	handlers map[uuid.UUID]core.Handler
}

func NewManager(client core.Client) Manager {
	return &managerImpl{
		Client:   client,
		handlers: make(map[uuid.UUID]core.Handler),
	}
}

func (m *managerImpl) Process(ctx context.Context, sessionID uuid.UUID) (core.Handler, error) {
	if h, err := m.getHandler(sessionID); err == nil {
		return h, nil
	}

	h, err := m.Client.Process(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	m.handlers[sessionID] = h
	return h, nil
}

func (m *managerImpl) ReleaseSession(ctx context.Context, sessionID uuid.UUID) error {
	err := m.Client.ReleaseSession(ctx, sessionID)
	if err != nil {
		return err
	}

	h, ok := m.handlers[sessionID]
	if ok {
		h.Close()
		delete(m.handlers, sessionID)
	}

	return nil
}

func (m *managerImpl) getHandler(sessionID uuid.UUID) (core.Handler, error) {
	const fn = "managerImpl.GetHandler"

	h, ok := m.handlers[sessionID]
	if !ok {
		return nil, fmt.Errorf("%s: handler for session %s not found; %w", fn, sessionID, ErrHandlerNotFound)
	}

	return h, nil
}
