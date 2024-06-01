package app

import "github.com/google/uuid"

type sessionKey struct {
	agent string
}

type sessionMap map[sessionKey]uuid.UUID
