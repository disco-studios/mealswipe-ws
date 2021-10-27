package codes

import (
	"log"
	"testing"
)

// TODO Make sure errors are of right type
func TestEncode(t *testing.T) {
	t.Run("decode yields encode input", func(t *testing.T) {
		input := 124315
		decoded := decodeRaw(encodeRaw(input))
		encodedEqualsDecoded(t, input, decoded)

		input = 0
		decoded = decodeRaw(encodeRaw(input))
		encodedEqualsDecoded(t, input, decoded)

		input = MAX_SESSION_CODE_RAW - 1
		decoded = decodeRaw(encodeRaw(input))
		encodedEqualsDecoded(t, input, decoded)
	})
}

func encodedEqualsDecoded(t *testing.T, input int, decoded int) {
	if input != decoded {
		log.Println("expected ", input, " got ", decoded)
		t.FailNow()
	}
}
