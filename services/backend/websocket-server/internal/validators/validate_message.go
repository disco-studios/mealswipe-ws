package validators

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/common/errors"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func ValidateMessage(userState *users.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if common.HasCreateMessage(genericMessage) {
		return ValidateMessageCreate(userState, genericMessage.GetCreateMessage())
	} else if common.HasJoinMessage(genericMessage) {
		return ValidateMessageJoin(userState, genericMessage.GetJoinMessage())
	} else if common.HasStartMessage(genericMessage) {
		return ValidateMessageStart(userState, genericMessage.GetStartMessage())
	} else if common.HasVoteMessage(genericMessage) {
		return ValidateMessageVote(userState, genericMessage.GetVoteMessage())
	} else {
		return &errors.UnknownWebsocketMessage{}
	}
}
