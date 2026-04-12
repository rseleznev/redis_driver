package translator

import (
	"errors"
	"strconv"
	"strings"

	"github.com/rseleznev/redis_driver/internal/models"
)

func (t Translator) Decode(buf []byte) (any, error) {
	dom := t.parse(buf)

	res := t.deserialize(dom)

	switch err := res.(type) {
	case error:
		return nil, err

	}
	
	return res, nil
}

// Parse парсит сырые данные и формирует корневой DOM-объект
func (t Translator) parse(input []byte) models.DOMPart {
	if len(input) == 0 {
		return models.DOMPart{}
	}
	
	idx, root := t.parsePart(0, input)
	idx++

	if root.ContentLen > 0 {
		for range root.ContentLen * 2 {
			offset, part := t.parsePart(idx, input)
			
			root.Content = append(root.Content, part)
			idx = offset + 1
		}	
	}

	return root
}

// parsePart парсит часть (элемент), принимает начальный индекс и срез, возвращает индекс,
// на котором остановился и прочитанную часть
func (t Translator) parsePart(index int, input []byte) (int, models.DOMPart) {
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

		index, part.ContentLen = t.parsePartLen(index, input)

		return index, part

	case '$': // Bulk strings
		part.PartType = "string"
		index++

		index, part.ValueLen = t.parsePartLen(index, input)
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

		index, part.ValueLen = t.parsePartLen(index, input)
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
		part.PartType = "error"
		str := []byte("ERR Unknown value type:")
		part.Value = append(part.Value, str...)

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
	}

	return index, part
}

// parsePartLen определяет длину элемента (может передаваться несколькими байтами)
func (t Translator) parsePartLen(index int, input []byte) (int, int) {
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

// deserialize десериализует DOM-объект в тип данных Go
func (t Translator) deserialize(domObj models.DOMPart) any {
	var result any

	switch domObj.PartType {
	case "string":
		return domObj.Value

	case "error":
		return t.deserializeError(domObj)

	case "map":
		m := map[string]string{}
		var key, value string

		for _, v := range domObj.Content {
			if key == "" {
				key = string(v.Value)
				continue
			}
			value = string(v.Value)
			m[key] = value

			key = ""
			value = ""
		}

		return m

	case "null":
		return models.ErrNoValue

	}

	return result
}

// deserializeError формирует тип error из объекта DOM
func (t Translator) deserializeError(domObj models.DOMPart) error {
	s := string(domObj.Value)
	strParts := strings.Fields(s)

	if strParts[1] == "Protocol" && strParts[2] == "error:" {
		errString := strParts[3] + strParts[4] + strParts[5] + strParts[6]
		err := errors.New(errString)

		return errors.Join(models.ErrRedisProtocol, err)
	}
	if strParts[1] == "Unknown" && strParts[2] == "value" && strParts[3] == "type:" {
		strParts = strParts[4:]
		var errString string
		for _, v := range strParts {
			errString = errString + v + " "
		}
		err := errors.New(errString)

		return errors.Join(models.ErrUnknownValueType, err)
	}

	return errors.New(s)
}