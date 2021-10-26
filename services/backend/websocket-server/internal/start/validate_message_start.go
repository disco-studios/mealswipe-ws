package start

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/common/errors"
	database "mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Start = []int16{constants.HostState_HOSTING}

func ValidateMessageStart(userState *users.UserState, startMessage *mealswipepb.StartMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := common.ValidateHostState(userState, AcceptibleHostStates_Start)
	if validateHostError != nil {
		return validateHostError
	}

	radiusValid, err := common.IsRadiusValid(startMessage.Radius)
	if err != nil {
		return err
	}
	if !radiusValid {
		return &errors.MessageValidationError{
			MessageType:   "start",
			Clarification: "invalid radius",
		}
	}

	latLonValid := common.LatLonWithinUnitedStates(startMessage.Lat, startMessage.Lng)
	if !latLonValid {
		return &errors.MessageValidationError{
			MessageType:   "start",
			Clarification: "invalid lat lng",
		}
	}

	sessionId, err := database.GetIdFromCode(userState.JoinedSessionCode)
	if err != nil || sessionId == "" {
		return err
	}
	if sessionId != userState.JoinedSessionId {
		return &errors.MessageValidationError{
			MessageType:   "start",
			Clarification: "session code links to session other than joined",
		}
	}

	return
}
