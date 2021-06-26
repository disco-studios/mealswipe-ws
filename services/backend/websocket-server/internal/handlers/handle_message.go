package handlers

import (
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessage(userState *core.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if core.HasCreateMessage(genericMessage) {
		return HandleMessageCreate(userState, genericMessage.GetCreateMessage())
	} else if core.HasJoinMessage(genericMessage) {
		return HandleMessageJoin(userState, genericMessage.GetJoinMessage())
	} else if core.HasStartMessage(genericMessage) {
		return HandleMessageStart(userState, genericMessage.GetStartMessage())
	} else if core.HasVoteMessage(genericMessage) {
		return HandleMessageVote(userState, genericMessage.GetVoteMessage())
	} else {
		return nil // TODO No message provided by ther user!! Figure out what to do here
	}
}
