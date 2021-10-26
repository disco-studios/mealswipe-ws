package join

// import (
// 	"log"
// 	"testing"

// 	"mealswipe.app/mealswipe/internal/business"
// 	"mealswipe.app/mealswipe/internal/constants"
// 	"mealswipe.app/mealswipe/internal/types"
// 	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
// )

// func TestValidateJoinMessage(t *testing.T) {
// 	userState := users.CreateUserState()
// 	joinMessage := &mealswipepb.JoinMessage{
// 		Nickname: "Cam the Man",
// 		Code:     "XCFHBB",
// 	}
// 	redisMock := business.LoadRedisMockClient()

// 	t.Run("HostState_JOINING invalid", func(t *testing.T) {
// 		userState.HostState = constants.HostState_JOINING
// 		if err := ValidateMessageJoin(userState, joinMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})
// 	t.Run("HostState_HOSTING invalid", func(t *testing.T) {
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateMessageJoin(userState, joinMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})
// 	t.Run("HostState_UNIDENTIFIED valid", func(t *testing.T) {
// 		redisMock.ExpectGet("code.XCFHBB").SetVal("b")
// 		userState.HostState = constants.HostState_UNIDENTIFIED
// 		if err := ValidateMessageJoin(userState, joinMessage); err != nil {
// 			log.Fatal(err)
// 		}
// 	})

// 	t.Run("Invalid nickname", func(t *testing.T) {
// 		joinMessage.Nickname = " Cam the Man"
// 		if err := ValidateMessageJoin(userState, joinMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})

// 	t.Run("Valid nickname", func(t *testing.T) {
// 		redisMock.ExpectGet("code.XCFHBB").SetVal("b")
// 		joinMessage.Nickname = "Cam the Man"
// 		if err := ValidateMessageJoin(userState, joinMessage); err != nil {
// 			log.Fatal(err)
// 		}
// 	})

// 	t.Run("Invalid code", func(t *testing.T) {
// 		joinMessage.Code = "ABCDEF"
// 		if err := ValidateMessageJoin(userState, joinMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})

// 	t.Run("Valid code but not running invalid", func(t *testing.T) {
// 		redisMock.ExpectGet("code.XCFHBB").RedisNil()
// 		joinMessage.Code = "XCFHBB"
// 		if err := ValidateMessageJoin(userState, joinMessage); err == nil {
// 			log.Fatal(err)
// 		}
// 	})

// 	t.Run("Valid code", func(t *testing.T) {
// 		redisMock.ExpectGet("code.XCFHBB").SetVal("b")
// 		joinMessage.Code = "XCFHBB"
// 		if err := ValidateMessageJoin(userState, joinMessage); err != nil {
// 			log.Fatal(err)
// 		}
// 	})
// }
