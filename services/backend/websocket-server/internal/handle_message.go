package mealswipe

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/create"
	"mealswipe.app/mealswipe/internal/join"
	"mealswipe.app/mealswipe/internal/start"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/internal/vote"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// TODO NULL SAFETY FROM PROTOBUF STUFF
func HandleMessage(userState *users.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if common.HasCreateMessage(genericMessage) {
		return create.HandleMessageCreate(userState, genericMessage.GetCreateMessage())
	} else if common.HasJoinMessage(genericMessage) {
		return join.HandleMessageJoin(userState, genericMessage.GetJoinMessage())
	} else if common.HasStartMessage(genericMessage) {
		return start.HandleMessageStart(userState, genericMessage.GetStartMessage())
	} else if common.HasVoteMessage(genericMessage) {
		return vote.HandleMessageVote(userState, genericMessage.GetVoteMessage())
	} else {
		return nil // TODO No message provided by ther user!! Figure out what to do here
	}
}
