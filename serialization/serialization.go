package serialization

import (
	"fmt"
	"log"
	"strconv"
)

// Decode принимает сырой набор байт и возвращает тип данных Go
func Decode(input []byte) any {
	switch input[0] {
	case '+': // Simple string
		result := decodeShortString(input[1:])

		return result

	case '%': // Maps
		for _, v := range input {
			fmt.Printf("Байт: %q \n", v)
		}
	
		result := decodeMap(input[1:])

		return result

	default:
		log.Fatal("Неизвестный тип данных в ответе")
	}
	
	return nil
}

// decodeShortString декодирует короткую строку до 10 символов
func decodeShortString(input []byte) string {
	filteredBytes := make([]byte, 0, 10)

	for _, v := range input {
		if v == '\r' || v == '\n' {
			continue
		}
		filteredBytes = append(filteredBytes, v)
	}

	return string(filteredBytes)
}

func decodeStringWithLen(input []byte, stringLen int) string {
	filteredBytes := make([]byte, 0, stringLen)

	for _, v := range input {
		if v == '\r' || v == '\n' {
			continue
		}
		filteredBytes = append(filteredBytes, v)
	}

	return string(filteredBytes)
}

// decodeMap декодирует map
func decodeMap(input []byte) map[string]string {

	// !В значении могут быть и строки и массивы в одной мапе!

	mapLen, _ := strconv.Atoi(string(input[0]))
	dataLen := len(input)
	result := make(map[string]string, mapLen) // точно ли можно делать строку ключом?

	var key string
	
	for i := 3; i < dataLen; {
		if input[i] != byte('$') {
			fmt.Println("Ключ/значение map не строка!")
			return nil
		}
		currElemLen, _ := strconv.Atoi(string(input[i+1]))
		nextElemOffset := 2 + currElemLen + 4
		if key == "" {
			key = decodeStringWithLen(input[i+2:i+nextElemOffset], currElemLen)
			i += nextElemOffset
			continue
		}
		value := decodeStringWithLen(input[i+2:nextElemOffset], currElemLen)
		result[key] = value

		i += nextElemOffset
	}

	return result
}