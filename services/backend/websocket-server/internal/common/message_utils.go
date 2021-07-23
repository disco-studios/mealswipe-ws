package common

import "mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"

func HasCreateMessage(genericMessage *mealswipepb.WebsocketMessage) bool {
	return (*genericMessage).GetCreateMessage() != nil
}

func HasStartMessage(genericMessage *mealswipepb.WebsocketMessage) bool {
	return (*genericMessage).GetStartMessage() != nil
}

func HasJoinMessage(genericMessage *mealswipepb.WebsocketMessage) bool {
	return (*genericMessage).GetJoinMessage() != nil
}

func HasVoteMessage(genericMessage *mealswipepb.WebsocketMessage) bool {
	return (*genericMessage).GetVoteMessage() != nil
}
