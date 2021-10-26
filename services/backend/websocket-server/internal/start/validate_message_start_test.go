package start

// import (
// 	"log"
// 	"testing"

// 	"mealswipe.app/mealswipe/internal/business"
// 	"mealswipe.app/mealswipe/internal/common/constants"
// 	"mealswipe.app/mealswipe/internal/users"
// 	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
// )

// func TestValidateStartMessage(t *testing.T) {
// 	userState := users.CreateUserState()
// 	validCode := "XCFHBB"
// 	userState.JoinedSessionCode = validCode
// 	userState.JoinedSessionId = "jsda"
// 	startMessage := &mealswipepb.StartMessage{
// 		Lat:    33.427204,   // Tempe, AZ, USA
// 		Lng:    -111.939896, // Tempe, AZ, USA
// 		Radius: 1000,
// 	}
// 	redisMock := business.LoadRedisMockClient()

// 	t.Run("HostState_UNIDENTIFIED invalid", func(t *testing.T) {
// 		if err := ValidateMessageStart(userState, startMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})
// 	t.Run("HostState_JOINING invalid", func(t *testing.T) {
// 		userState.HostState = constants.HostState_JOINING
// 		if err := ValidateMessageStart(userState, startMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})
// 	t.Run("HostState_HOSTING valid", func(t *testing.T) {
// 		redisMock.ExpectGet("code." + userState.JoinedSessionCode).SetVal(userState.JoinedSessionId)
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateMessageStart(userState, startMessage); err != nil {
// 			log.Fatal(err)
// 		}
// 	})

// 	t.Run("Bloomington, MN, USA valid", func(t *testing.T) {
// 		redisMock.ExpectGet("code." + userState.JoinedSessionCode).SetVal(userState.JoinedSessionId)
// 		startMessage.Lat = 44.84079
// 		startMessage.Lng = -93.298279
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateMessageStart(userState, startMessage); err != nil {
// 			t.FailNow()
// 		}
// 	})
// 	t.Run("Mexico City, MX invalid", func(t *testing.T) {
// 		startMessage.Lat = 19.4326
// 		startMessage.Lng = 99.1332
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateMessageStart(userState, startMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})
// 	t.Run("Zero lat lng invalid", func(t *testing.T) {
// 		startMessage.Lat = 0
// 		startMessage.Lng = 0
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateMessageStart(userState, startMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})

// 	t.Run("Session started fails", func(t *testing.T) {
// 		redisMock.ExpectGet("code." + userState.JoinedSessionCode).RedisNil()
// 		startMessage.Lat = 44.84079
// 		startMessage.Lng = -93.298279
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateMessageStart(userState, startMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})

// 	t.Run("Session exists and is correct", func(t *testing.T) {
// 		redisMock.ExpectGet("code." + userState.JoinedSessionCode).SetVal(userState.JoinedSessionId)
// 		startMessage.Lat = 44.84079
// 		startMessage.Lng = -93.298279
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateMessageStart(userState, startMessage); err != nil {
// 			log.Fatal(err)
// 		}
// 	})

// 	t.Run("Session exists and is incorrect fails", func(t *testing.T) {
// 		redisMock.ExpectGet("code." + userState.JoinedSessionCode).SetVal(userState.JoinedSessionId + "f")
// 		startMessage.Lat = 44.84079
// 		startMessage.Lng = -93.298279
// 		userState.HostState = constants.HostState_HOSTING
// 		if err := ValidateMessageStart(userState, startMessage); err == nil {
// 			t.FailNow()
// 		}
// 	})
// }
