package common

// import (
// 	"log"
// 	"testing"

// 	"mealswipe.app/mealswipe/internal/constants"
// 	"mealswipe.app/mealswipe/internal/types"
// )

// // TODO Make sure errors are of right type
// func TestValidateCreateMessage(t *testing.T) {
// 	userState := users.CreateUserState()

// 	t.Run("valid in list", func(t *testing.T) {
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateHostState(userState, []int16{
// 			constants.HostState_HOSTING,
// 			constants.HostState_JOINING,
// 		}); err != nil {
// 			log.Fatal(err)
// 		}
// 	})

// 	t.Run("valid single", func(t *testing.T) {
// 		userState.HostState = constants.HostState_UNIDENTIFIED
// 		if err := ValidateHostState(userState, []int16{
// 			constants.HostState_UNIDENTIFIED,
// 		}); err != nil {
// 			log.Fatal(err)
// 		}
// 	})

// 	t.Run("invalid in list", func(t *testing.T) {
// 		userState.HostState = constants.HostState_UNIDENTIFIED
// 		if err := ValidateHostState(userState, []int16{
// 			constants.HostState_HOSTING,
// 			constants.HostState_JOINING,
// 		}); err == nil {
// 			t.FailNow()
// 		}
// 	})

// 	t.Run("invalid single", func(t *testing.T) {
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateHostState(userState, []int16{
// 			constants.HostState_UNIDENTIFIED,
// 		}); err == nil {
// 			t.FailNow()
// 		}
// 	})
// }
