package validators

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/common/errors"
	"mealswipe.app/mealswipe/internal/create"
	"mealswipe.app/mealswipe/internal/join"
	"mealswipe.app/mealswipe/internal/start"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/internal/vote"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// TODO Check for empty states
func ValidateMessage(userState *users.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if common.HasCreateMessage(genericMessage) {
		return create.ValidateMessage(userState, genericMessage.GetCreateMessage())
	} else if common.HasJoinMessage(genericMessage) {
		return join.ValidateMessage(userState, genericMessage.GetJoinMessage())
	} else if common.HasStartMessage(genericMessage) {
		return start.ValidateMessage(userState, genericMessage.GetStartMessage())
	} else if common.HasVoteMessage(genericMessage) {
		return vote.ValidateMessage(userState, genericMessage.GetVoteMessage())
	} else {
		return &errors.UnknownWebsocketMessage{}
	}
}
