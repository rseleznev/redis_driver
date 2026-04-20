package translator

import (
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

func (t *Translator) DecodeWithProceeding(input []byte) bool {
	// закидываем срез в структуру
	t.setDecodingData(input)

	// декодируем с начала или продолжаем
	if t.isDecodeProceeding() {
		// декодирование продолжается не с начала
		// продолжаем декодировать t.decodingDOMPart
		// return
	}

	var idx int

	// декодируем с начала
	for {
		idx = t.parsePartNew(idx)
		if idx >= t.decodingDataLen() {
			break
		}
	}

	// указываем, что нужно будет продолжить декодирование
	t.setDecodeProceeding()

	return true
}

func (t *Translator) parsePartNew(idx int) int {
	var finished bool
	var part models.DOMPart

	// определяем тип объекта
	switch t.decodingData[idx] {
	case '+': // Simple string 
		idx, finished, part = t.parseSimpleString(idx)

	case '$': // Bulk strings
		idx, finished, part = t.parseBulkString(idx)
	
	case '%': // Maps
		// парсим map

	case ':': // Integers
		// парсим число
	
	case '*': // Arrays
		// парсим массив

	case '_': // Nil
		// парсим nil

	case '-': // Simple Errors
		// парсим ошибку

	default:
		panic("unsupported RESP3 data type")
	}

	if !finished {
		t.decodingDOMPart = part
	} else {
		t.decodedDOM = part // будет добавляться по-разному в зависимости от типа!
	}

	return idx
}

func (t *Translator) parseSimpleString(idx int) (int, bool, models.DOMPart) {
	var simpleString models.DOMPart

	simpleString.PartType = "string"
	idx++

	for {
		if t.isDataEnd(idx) {
			return idx, false, simpleString
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnd(idx) {
				return idx, false, simpleString
			}

			if t.decodingData[idx] == '\n' {
				break
			}
		}

		simpleString.Value = append(simpleString.Value, t.decodingData[idx])
		idx++
	}
	simpleString.ValueLen = len(simpleString.Value)
	
	return idx, true, simpleString
}

func (t *Translator) parseBulkString(idx int) (int, bool, models.DOMPart) {
	var bulkString models.DOMPart
	var finished bool

	bulkString.PartType = "string"
	idx++

	if t.isDataEnd(idx) {
		return idx, false, bulkString
	}
	idx, finished, bulkString.ValueLenBytes = t.parsePartLenNew(idx)
	
	if finished {
		lenString := string(bulkString.ValueLenBytes)
		bulkString.ValueLen, _ = strconv.Atoi(lenString)
	} else {
		return idx, false, bulkString
	}
	idx++

	for {
		if t.isDataEnd(idx) {
			return idx, false, bulkString
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnd(idx) {
				return idx, false, bulkString
			}

			if t.decodingData[idx] == '\n' {
				break
			}
		}

		bulkString.Value = append(bulkString.Value, t.decodingData[idx])
		idx++
	}

	return idx, true, bulkString
}

func (t *Translator) parsePartLenNew(idx int) (int, bool, []byte) {
	valueLenBytes := make([]byte, 0, 5)
	idx++
	
	for {
		if t.isDataEnd(idx) {
			return idx, false, valueLenBytes
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnd(idx) {
				return idx, false, valueLenBytes
			}

			if t.decodingData[idx] == '\n' {
				break
			}
		}

		valueLenBytes = append(valueLenBytes, t.decodingData[idx])
		idx++
	}

	return idx, true, valueLenBytes
}