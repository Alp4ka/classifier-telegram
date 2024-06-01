package app

import (
	"github.com/Alp4ka/classifier-telegram/internal/interactions/core"
)

type Config struct {
	// Security.
	APIKey string

	// Core client.
	CoreClient core.Client
}
