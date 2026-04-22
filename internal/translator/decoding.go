package translator

import (
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

func (t *Translator) Decode(input []byte) any {
	// закидываем срез в структуру
	t.setDecodingData(input)

	var idx int

	_, result := t.parsePart(idx)

	return result
}

// parsePart парсит "объект", в случае команд PING, SET и GET это будут одиночные строки
// в случае команды HELLO 3 будет собираться map[string]string
func (t *Translator) parsePart(idx int) (int, any) {

	// определяем тип объекта
	switch t.decodingData[idx] {
	case '+': // Simple string 
		return t.parseSimpleString(idx)

	case '$': // Bulk strings
		return t.parseBulkString(idx)
	
	case '%': // Maps
		return t.parseMap(idx)

	case ':': // Integers
		return t.parseInteger(idx)
	
	case '*': // Arrays
		return t.parseArray(idx)

	case '_': // Nil
		idx += 2	
		return idx, models.ErrNoValue

	case '-': // Simple Errors
		// парсим ошибку

	default:
		panic("unsupported RESP3 data type")
	}

	return idx, nil
}

func (t *Translator) parseSimpleString(idx int) (int, any) {
	str := make([]byte, 0, 10)
	idx++

	for {
		if t.isDataEnded(idx) {
			return idx, nil
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnded(idx) {
				return idx, nil
			}

			if t.decodingData[idx] == '\n' {
				break
			}
		}

		str = append(str, t.decodingData[idx])
		idx++
	}
	
	return idx, str
}

func (t *Translator) parseBulkString(idx int) (int, any) {
	var strLen int
	idx++

	if t.isDataEnded(idx) {
		return idx, nil
	}
	idx, strLen = t.parsePartLen(idx)
	str := make([]byte, 0, strLen)
	idx++

	for strLen > 0 {
		if t.isDataEnded(idx) {
			return idx, str
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnded(idx) {
				return idx, str
			}

			if t.decodingData[idx] == '\n' {
				break
			}
		}

		str = append(str, t.decodingData[idx])
		idx++
	}

	return idx, str
}

func (t *Translator) parseMap(idx int) (int, any) {
	var mapLen int
	idx++

	if t.isDataEnded(idx) {
		return idx, nil
	}
	idx, mapLen = t.parsePartLen(idx)

	m := make(map[string]string, mapLen)
	var key, value string

	for mapLen > 0 {
		idx++
		if t.isDataEnded(idx) {
			return idx, m
		}

		idx, key, value = t.parseMapKeyAndValue(idx)
		m[key] = value

		mapLen--
	}

	return idx, m
}

func (t *Translator) parseMapKeyAndValue(idx int) (int, string, string) {
	var res any

	idx, res = t.parsePart(idx)
	key := res.(string)
	idx++

	idx, res = t.parsePart(idx)
	value := res.(string)

	return idx, key, value
}

func (t *Translator) parseInteger(idx int) (int, any) {
	integer := make([]byte, 0, 3)
	idx++

	for {
		if t.isDataEnded(idx) {
			return idx, integer
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnded(idx) {
				return idx, integer
			}

			if t.decodingData[idx] == '\n' {
				break
			}
		}

		integer = append(integer, t.decodingData[idx])
		idx++
	}

	return idx, integer
}

func (t *Translator) parseArray(idx int) (int, any) {
	var arrLen int
	var arrPart any
	idx++

	if t.isDataEnded(idx) {
		return idx, nil
	}

	idx, arrLen = t.parsePartLen(idx)
	arr := make([]any, 0, arrLen)

	for arrLen > 0 {
		idx++
		if t.isDataEnded(idx) {
			return idx, arr
		}

		idx, arrPart = t.parsePart(idx)
		arr = append(arr, arrPart)
		arrLen--
	}

	return idx, arr
}

func (t *Translator) parsePartLen(idx int) (int, int) {
	valueLenBytes := make([]byte, 0, 5)
	idx++
	
	for {
		if t.isDataEnded(idx) {
			return idx, 0
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnded(idx) {
				return idx, 0
			}

			if t.decodingData[idx] == '\n' {
				break
			}
		}

		valueLenBytes = append(valueLenBytes, t.decodingData[idx])
		idx++
	}
	lenString := string(valueLenBytes)
	lenResult, _ := strconv.Atoi(lenString)

	return idx, lenResult
}