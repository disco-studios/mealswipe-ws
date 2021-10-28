package messages

import (
	"fmt"

	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/messages/create"
	"mealswipe.app/mealswipe/internal/messages/join"
	"mealswipe.app/mealswipe/internal/messages/start"
	"mealswipe.app/mealswipe/internal/messages/vote"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// TODO NULL SAFETY FROM PROTOBUF STUFF
func HandleMessage(userState *types.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if common.HasCreateMessage(genericMessage) {
		err = create.HandleMessage(userState, genericMessage.GetCreateMessage())
		if err != nil {
			err = fmt.Errorf("handle create message: %w", err)
		}
		return
	} else if common.HasJoinMessage(genericMessage) {
		err = join.HandleMessage(userState, genericMessage.GetJoinMessage())
		if err != nil {
			err = fmt.Errorf("handle join message: %w", err)
		}
		return
	} else if common.HasStartMessage(genericMessage) {
		err = start.HandleMessage(userState, genericMessage.GetStartMessage())
		if err != nil {
			err = fmt.Errorf("handle start message: %w", err)
		}
		return
	} else if common.HasVoteMessage(genericMessage) {
		err = vote.HandleMessage(userState, genericMessage.GetVoteMessage())
		if err != nil {
			err = fmt.Errorf("handle vote message: %w", err)
		}
		return
	} else {
		return nil // TODO No message provided by ther user!! Figure out what to do here
	}
}
