package grpc

import (
	"fmt"

	api "github.com/Alp4ka/classifier-api"
	"github.com/Alp4ka/classifier-telegram/internal/interactions/core"
)

func protoActionToAction(actionProto api.Action) (core.Action, error) {
	switch actionProto {
	case api.Action_ACTION_FINISH:
		return core.ActionFinish, nil
	case api.Action_ACTION_LISTEN:
		return core.ActionListen, nil
	case api.Action_ACTION_RESPOND:
		return core.ActionRespond, nil
	}
	return "", fmt.Errorf("cannot recognize proto action %d; %w", actionProto, core.ErrUnknownAction)
}
