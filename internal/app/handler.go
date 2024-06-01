package app

import (
	"context"
	"errors"
	"strconv"

	classifiertelegram "github.com/Alp4ka/classifier-telegram"
	"github.com/Alp4ka/classifier-telegram/internal/interactions/core"
	"github.com/Alp4ka/mlogger"
	"github.com/Alp4ka/mlogger/field"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
)

func (a *App) handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	a.mu.Lock()
	defer a.mu.Unlock()

	agent := strconv.FormatInt(update.Message.Chat.ID, 10)

	// TODO: async would not work properly. mutex to handler.
	var (
		sessionID uuid.UUID
		err       error
	)
	session := a.cache.Get(agent)
	if session == nil || session.IsExpired() {
		sessionID, err = a.client.AcquireSession(ctx, agent, classifiertelegram.AppName)
		if err != nil {
			mlogger.L(ctx).Error("failed to acquire session", field.Error(err))
			return
		}
		a.cache.Set(agent, sessionID, ttlcache.DefaultTTL)
	} else {
		sessionID = session.Value()
	}

cycle:
	if update.Message != nil {
		handler, err := a.client.Process(ctx, sessionID)
		if err != nil {
			mlogger.L(ctx).Error("failed to process", field.Error(err))
			return
		}

		if handler.GetOutputHandler().HasMessage() {
			output, err := handler.GetOutputHandler().Await()
			if err != nil {
				err = errors.Join(err, handler.Error())
				mlogger.L(ctx).Error("failed to retrieve message from output handler", field.Error(err))
				return
			}

			switch output.Action {
			case core.ActionListen:
				goto cycle
			case core.ActionFinish:
				a.cache.Delete(agent)
				return
			case core.ActionRespond:
				err = a.respond(ctx, b, update, output.UserResponse)
				if err != nil {
					mlogger.L(ctx).Error("failed to respond", field.Error(err))
				}
				goto cycle
			}
		}

		err = handler.GetInputHandler().Handle(&core.ProcessInput{
			UserInput: update.Message.Text,
			RequestID: uuid.New(),
		})
		if err != nil {
			err = errors.Join(err, handler.Error())
			mlogger.L(ctx).Error("failed to send message to input handler", field.Error(err))
			return
		}

		output, err := handler.GetOutputHandler().Await()
		if err != nil {
			mlogger.L(ctx).Error("failed to send message to input handler", field.Error(err))
			err = errors.Join(err, handler.Error())
			return
		}

		switch output.Action {
		case core.ActionListen:
			return
		case core.ActionFinish:
			a.cache.Delete(agent)
			return
		case core.ActionRespond:
			err = a.respond(ctx, b, update, output.UserResponse)
			if err != nil {
				mlogger.L(ctx).Error("failed to respond after handle", field.Error(err))
			}
			return
		}
	}
}

func (a *App) respond(ctx context.Context, b *bot.Bot, update *models.Update, text string) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   text,
	})

	const maxRetries = 5
	for cnt := 0; cnt < maxRetries && err != nil; cnt++ {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   text,
		})
	}

	return err
}
