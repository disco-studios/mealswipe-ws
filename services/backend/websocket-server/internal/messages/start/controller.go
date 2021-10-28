package start

import (
	"fmt"

	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/pkg/mealswipe"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Start = []int16{mealswipe.HostState_HOSTING}

func HandleMessage(userState *types.UserState, startMessage *mealswipepb.StartMessage) (err error) {
	err = sessions.Start(userState.JoinedSessionCode, userState.JoinedSessionId, startMessage.Lat, startMessage.Lng, startMessage.Radius, startMessage.CategoryId)
	if err != nil {
		err = fmt.Errorf("start session: %w", err)
		return
	}

	err = userState.PubsubWebsocketResponse(&mealswipepb.WebsocketResponse{
		GameStartedMessage: &mealswipepb.GameStartedMessage{},
	})
	if err != nil {
		err = fmt.Errorf("send game start message: %w", err)
		return
	}
	return
}

func ValidateMessage(userState *types.UserState, startMessage *mealswipepb.StartMessage) (err error) {
	// Validate that the user is in a state that can do this action
	err = common.ValidateHostState(userState, AcceptibleHostStates_Start)
	if err != nil {
		err = fmt.Errorf("validate host state: %w", err)
		return
	}

	radiusValid, err := common.IsRadiusValid(startMessage.Radius)
	if err != nil {
		err = fmt.Errorf("validate radius: %w", err)
		return
	}
	if !radiusValid {
		return &mealswipe.MessageValidationError{
			MessageType:   "start",
			Clarification: "invalid radius",
		}
	}

	latLonValid := common.LatLonWithinUnitedStates(startMessage.Lat, startMessage.Lng)
	if !latLonValid {
		return &mealswipe.MessageValidationError{
			MessageType:   "start",
			Clarification: "invalid lat lng",
		}
	}

	sessionId, err := sessions.GetIdFromCode(userState.JoinedSessionCode)
	if err != nil || sessionId == "" {
		err = fmt.Errorf("get id from code: %w", err)
		return err
	}
	if sessionId != userState.JoinedSessionId {
		return &mealswipe.MessageValidationError{
			MessageType:   "start",
			Clarification: "session code links to session other than joined",
		}
	}

	return
}
