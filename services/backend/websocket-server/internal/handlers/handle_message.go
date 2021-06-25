package handlers

import (
	"mealswipe.app/mealswipe/core"
	"mealswipe.app/mealswipe/mealswipepb"
)

func HandleMessage(userState *core.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if core.HasCreateMessage(genericMessage) {
		return HandleMessageCreate(userState, genericMessage.GetCreateMessage())
	} else if core.HasJoinMessage(genericMessage) {
		return HandleMessageJoin(userState, genericMessage.GetJoinMessage())
	} else if core.HasStartMessage(genericMessage) {
		return HandleMessageStart(userState, genericMessage.GetStartMessage()) // TODO Add message response
	} else if core.HasVoteMessage(genericMessage) {
		// TODO: Next thing to do here: Implement voting and win conditions
		return HandleMessageVote(userState, genericMessage.GetVoteMessage()) // TODO Add message response
	} else {
		return nil // TODO No message provided by ther user!! Figure out what to do here
	}
}
