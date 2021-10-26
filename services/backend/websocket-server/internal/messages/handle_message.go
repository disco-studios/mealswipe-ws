package messages

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/messages/create"
	"mealswipe.app/mealswipe/internal/messages/join"
	"mealswipe.app/mealswipe/internal/messages/start"
	"mealswipe.app/mealswipe/internal/messages/vote"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// TODO NULL SAFETY FROM PROTOBUF STUFF
func HandleMessage(userState *users.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if common.HasCreateMessage(genericMessage) {
		return create.HandleMessage(userState, genericMessage.GetCreateMessage())
	} else if common.HasJoinMessage(genericMessage) {
		return join.HandleMessage(userState, genericMessage.GetJoinMessage())
	} else if common.HasStartMessage(genericMessage) {
		return start.HandleMessage(userState, genericMessage.GetStartMessage())
	} else if common.HasVoteMessage(genericMessage) {
		return vote.HandleMessage(userState, genericMessage.GetVoteMessage())
	} else {
		return nil // TODO No message provided by ther user!! Figure out what to do here
	}
}
