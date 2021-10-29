package messages

import (
	"fmt"

	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/messages/create"
	"mealswipe.app/mealswipe/internal/messages/join"
	"mealswipe.app/mealswipe/internal/messages/start"
	"mealswipe.app/mealswipe/internal/messages/vote"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// TODO NULL SAFETY FROM PROTOBUF STUFF
func HandleMessage(userState *types.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	logger := logging.Get()
	if common.HasCreateMessage(genericMessage) {
		logger.Info("message received", logging.Metric("message_received"), zap.String("type", "create"))
		err = create.HandleMessage(userState, genericMessage.GetCreateMessage())
		if err != nil {
			err = fmt.Errorf("handle create message: %w", err)
		}
		return
	} else if common.HasJoinMessage(genericMessage) {
		logger.Info("message received", logging.Metric("message_received"), zap.String("type", "join"))
		err = join.HandleMessage(userState, genericMessage.GetJoinMessage())
		if err != nil {
			err = fmt.Errorf("handle join message: %w", err)
		}
		return
	} else if common.HasStartMessage(genericMessage) {
		logger.Info("message received", logging.Metric("message_received"), zap.String("type", "start"))
		err = start.HandleMessage(userState, genericMessage.GetStartMessage())
		if err != nil {
			err = fmt.Errorf("handle start message: %w", err)
		}
		return
	} else if common.HasVoteMessage(genericMessage) {
		logger.Info("message received", logging.Metric("message_received"), zap.String("type", "vote"))
		err = vote.HandleMessage(userState, genericMessage.GetVoteMessage())
		if err != nil {
			err = fmt.Errorf("handle vote message: %w", err)
		}
		return
	} else {
		return nil // TODO No message provided by ther user!! Figure out what to do here
	}
}
