package common

import "github.com/Treble-Development/mealswipe-proto/mealswipepb"

func HasCreateMessage(genericMessage *mealswipepb.WebsocketMessage) bool {
	return (*genericMessage).GetCreateMessage() != nil
}

func HasStartMessage(genericMessage *mealswipepb.WebsocketMessage) bool {
	return (*genericMessage).GetStartMessage() != nil
}

func HasJoinMessage(genericMessage *mealswipepb.WebsocketMessage) bool {
	return (*genericMessage).GetJoinMessage() != nil
}

func HasRejoinMessage(genericMessage *mealswipepb.WebsocketMessage) bool {
	return (*genericMessage).GetRejoinMessage() != nil
}

func HasVoteMessage(genericMessage *mealswipepb.WebsocketMessage) bool {
	return (*genericMessage).GetVoteMessage() != nil
}

func HasLobbyInfoMessage(genericMessage *mealswipepb.WebsocketResponse) bool {
	return (*genericMessage).GetLobbyInfoMessage() != nil
}
