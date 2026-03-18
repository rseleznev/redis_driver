package message

import (
	"fmt"
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

// Parse парсит сырые данные и формирует корневой DOM-объект
func Parse(input []byte) models.DOMPart {
	if len(input) == 0 {
		return models.DOMPart{}
	}
	
	idx, root := parsePart(0, input)
	idx++

	if root.ContentLen > 0 {
		for range root.ContentLen * 2 {
			offset, part := parsePart(idx, input)
			
			root.Content = append(root.Content, part)
			idx = offset + 1
		}	
	}

	return root
}

// parsePart парсит часть (элемент), принимает начальный индекс и срез, возвращает индекс,
// на котором остановился и прочитанную часть
func parsePart(index int, input []byte) (int, models.DOMPart) {
	var part models.DOMPart
	var partValue []byte

	switch input[index] {
	case '+': // Simple string
		part.PartType = "string"
		index++

		for {
			if input[index] == '\r' {
				if input[index+1] == '\n' {
					index++
					break
				}
			}
			partValue = append(partValue, input[index])
			index++
		}
		part.ValueLen = len(partValue)
		part.Value = append(part.Value, partValue...)

		return index, part

	case '%': // Maps
		part.PartType = "map"
		index++

		index, part.ContentLen = parsePartLen(index, input)

		return index, part

	case '$': // Bulk strings
		part.PartType = "string"
		index++

		index, part.ValueLen = parsePartLen(index, input)
		toFill := part.ValueLen
		index++

		for toFill > 0 {
			if input[index] == '\r' {
				index++
				continue
			}
			if input[index] == '\n' {
				index++
				continue
			}
			partValue = append(partValue, input[index])
			index++
			toFill--
		}
		part.Value = append(part.Value, partValue...)

		if input[index] == '\r' {
			if input[index+1] == '\n' {
				index++
				return index, part
			}
		}

	case ':': // Integers
		part.PartType = "int"
		index++

		for {
			if input[index] == '\r' {
				if input[index+1] == '\n' {
					index++
					break
				}
			}
			partValue = append(partValue, input[index])
			index++
		}
		part.ValueLen = len(partValue)
		part.Value = append(part.Value, partValue...)

		return index, part
	
	case '*': // Arrays
		part.PartType = "array"
		index++

		index, part.ValueLen = parsePartLen(index, input)
		part.ContentLen = part.ValueLen

	case '_': // Nil
		part.PartType = "null"
		index += 2

		return index, part

	case '-': // Simple Errors
		part.PartType = "error"
		index++

		for {
			if input[index] == '\r' {
				if input[index+1] == '\n' {
					index++
					break
				}
			}
			partValue = append(partValue, input[index])
			index++
		}
		part.ValueLen = len(partValue)
		part.Value = append(part.Value, partValue...)

		return index, part

	default:
		fmt.Println("Неизвестный тип данных")
	}

	return index, part
}

// parsePartLen определяет длину элемента (может передаваться несколькими байтами)
func parsePartLen(index int, input []byte) (int, int) {
	var partLen int
	var lenBytes []byte

	for {
		if input[index] == '\r' {
			if input[index+1] == '\n' {
				index++
				break
			}
		}
		lenBytes = append(lenBytes, input[index])
		index++
	}
	lenString := string(lenBytes)
	partLen, _ = strconv.Atoi(lenString)

	return index, partLen
}