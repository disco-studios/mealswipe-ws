package codes

import (
	"math"
	"math/rand"
	"time"
)

// vowels removed to reduce odds of bad words
var SESSION_CODE_CHARSET []string = []string{
	"B", "C", "D", "F", "G", "H", "J",
	"K", "L", "M", "N", "P", "Q", "R",
	"S", "T", "V", "W", "X", "Y", "Z",
}

const SESSION_CODE_LENGTH int = 6

var SESSION_CODE_BASE = len(SESSION_CODE_CHARSET)
var MAX_SESSION_CODE_RAW int = int(math.Pow(
	float64(SESSION_CODE_BASE),
	float64(SESSION_CODE_LENGTH),
))

func GenerateRandomRaw() int {
	randSource := rand.NewSource(time.Now().UnixNano())
	return rand.New(randSource).Intn(MAX_SESSION_CODE_RAW)
}

func EncodeRaw(rawCode int) string {
	out := ""
	for i := 0; i < SESSION_CODE_LENGTH; i++ {
		out = SESSION_CODE_CHARSET[rawCode%SESSION_CODE_BASE] + out
		rawCode = rawCode / SESSION_CODE_BASE
	}
	return out
}

func DecodeRaw(code string) int {
	out := 0
	// Go through each digit
	for _, codeChar := range code {
		out *= SESSION_CODE_BASE
		out += findValueOfCodeChar(string(codeChar))
	}
	return out
}

func findValueOfCodeChar(codeChar string) int {
	for charInd, charsetChar := range SESSION_CODE_CHARSET {
		if charsetChar == codeChar {
			return charInd
		}
	}
	panic("Could not find char for int")
}
