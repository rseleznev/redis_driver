package serialization

import (
	"fmt"
	"log"
	"strconv"

	"github.com/rseleznev/redis_driver/models"
)

// Decode принимает сырой набор байт и возвращает тип данных Go
func Decode(input []byte) any {
	switch input[0] {
	case '+': // Simple string
		result := decodeShortString(input[1:])

		return result

	case '%': // Maps
		result := decodeMap(input[1:])

		testResult := parseMap(input[1:])
		fmt.Println(testResult)

		return result

	default:
		log.Fatal("Неизвестный тип данных в ответе")
	}
	
	return nil
}

// decodeShortString декодирует короткую строку
func decodeShortString(input []byte) string {
	filteredBytes := make([]byte, 0, 40)

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
	for i, v := range input {
		fmt.Printf("Байт: %q, индекс: %d \n", v, i)
	}

	// !В значении могут быть и строки и массивы в одной мапе!

	mapLen, _ := strconv.Atoi(string(input[0]))
	dataLen := len(input)
	result := make(map[string]string, mapLen) // точно ли можно делать строку ключом?
	
	for i := 3; i < dataLen; {
		keyType := input[i]
		keyLen, _ := strconv.Atoi(string(input[i+1]))

		valueType := input[i+keyLen+4+2]
		valueLen, _ := strconv.Atoi(string(input[i+keyLen+4+2+1]))

		// Временно только строки
		if keyType != valueType && keyType == byte('$') {
			fmt.Println("Ключ/значение map не строка!")
			return result
		}
	
		keyStartOffset := i+2
		keyEndOffset := i+keyLen+4+2
		valueStartOffset := keyEndOffset+2
		valueEndOffset := valueStartOffset+valueLen+4

		// fmt.Printf("Индекс начала ключа: %d, индекс кончика ключа: %d, индекс начала значения: %d, индекс кончика значения: %d \n",
		// 	keyStartOffset, keyEndOffset, valueStartOffset, valueEndOffset)

		key := decodeStringWithLen(input[keyStartOffset:keyEndOffset], keyLen)
		value := decodeStringWithLen(input[valueStartOffset:valueEndOffset], valueLen)

		result[key] = value

		i = valueEndOffset
	}

	return result
}

func parseMap(input []byte) models.ParsedPesponse {
	mapLen, _ := strconv.Atoi(string(input[0]))
	
	root := models.ParsedPesponse{
		MainType: "map",
		TotalLen: mapLen,
	}

outer:
	for i := 3; i < len(input); {
		if input[i] == '\r' {
			i++
			continue
		}
		if input[i] == '\n' {
			i++
			continue
		}
		
		var partType string
		var partLen int

		switch input[i] {
		case '$':
			partType = "string"
			partLen, _ = strconv.Atoi(string(input[i+1])) // !длина может быть двузначной!

			i += 2

		case ':':
			partType = "int"

		case '*':
			partType = "array"
			partLen, _ = strconv.Atoi(string(input[i+1]))
		default:
			break outer
		}
		
		part := models.ParsedResponsePart{
			PartType: partType,
			Len: partLen,
		}

		// Парсим тип, у которого длина > 0
		// В случае числа цикл не должен запуститься
		for toFill := partLen; toFill > 0; {
			if input[i] == '\r' {
				i++
				continue
			}
			if input[i] == '\n' {
				i++
				continue
			}
			part.Data = append(part.Data, input[i])
			toFill--
			i++
		}

		if partType == "int" {
			i++
			for {
				if input[i] == '\r' {
					break
				}
				if input[i] == '\n' {
					break
				}
				fmt.Printf("Зашел в цикл числа: %q \n", input[i])
				part.Data = append(part.Data, input[i])
				i++
			}
		}

		root.Data = append(root.Data, part)
	}

	return root
}