package validators

import (
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func ValidateMessage(userState *core.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if core.HasCreateMessage(genericMessage) {
		return ValidateMessageCreate(userState, genericMessage.GetCreateMessage())
	} else if core.HasJoinMessage(genericMessage) {
		return ValidateMessageJoin(userState, genericMessage.GetJoinMessage())
	} else if core.HasStartMessage(genericMessage) {
		return ValidateMessageStart(userState, genericMessage.GetStartMessage())
	} else if core.HasVoteMessage(genericMessage) {
		return ValidateMessageVote(userState, genericMessage.GetVoteMessage())
	} else {
		return &core.UnknownWebsocketMessage{}
	}
}
