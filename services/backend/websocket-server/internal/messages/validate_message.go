package messages

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/messages/create"
	"mealswipe.app/mealswipe/internal/messages/join"
	"mealswipe.app/mealswipe/internal/messages/start"
	"mealswipe.app/mealswipe/internal/messages/vote"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/pkg/mealswipe"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// TODO Check for empty states
func ValidateMessage(userState *types.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if common.HasCreateMessage(genericMessage) {
		return create.ValidateMessage(userState, genericMessage.GetCreateMessage())
	} else if common.HasJoinMessage(genericMessage) {
		return join.ValidateMessage(userState, genericMessage.GetJoinMessage())
	} else if common.HasStartMessage(genericMessage) {
		return start.ValidateMessage(userState, genericMessage.GetStartMessage())
	} else if common.HasVoteMessage(genericMessage) {
		return vote.ValidateMessage(userState, genericMessage.GetVoteMessage())
	} else {
		return &mealswipe.UnknownWebsocketMessage{}
	}
}
