package mealswipe

import "fmt"

type InvalidHostStateError struct {
	Allowed  []int16
	Received int16
}

func (e *InvalidHostStateError) Error() string {
	return fmt.Sprint(
		"invalid host state: expected ",
		e.Allowed,
		" had ",
		e.Received,
	)
}

type MessageValidationError struct {
	MessageType   string
	Clarification string
}

func (e *MessageValidationError) Error() string {
	return fmt.Sprint(
		"validation error: ",
		e.MessageType,
		": ",
		e.Clarification,
	)
}

type UnknownWebsocketMessage struct {
}

func (e *UnknownWebsocketMessage) Error() string {
	return "unknown websocket message"
}

type KeyExistsError struct {
}

func (e *KeyExistsError) Error() string {
	return "existing key already exists"
}
