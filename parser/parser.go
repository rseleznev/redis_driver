package parser

import (
	"fmt"
)

func ParseResponse(input []byte) string {
	var result string

	for _, v := range input {
		fmt.Printf("Байт: %q \n", v)
		result += string(v)
	}

	return result
}