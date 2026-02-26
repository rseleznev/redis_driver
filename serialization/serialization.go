package serialization

import (
	"fmt"
)

func Decode(input []byte) string {
	var result string

	for _, v := range input {
		fmt.Printf("Байт: %q \n", v)
		result += string(v)
	}

	return result
}