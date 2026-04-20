package translator

import "github.com/rseleznev/redis_driver/internal/models"

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
		// парсим строку
	
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
		t.decodingDOMPart = append(t.decodingDOMPart, part)
	} else {
		t.decodedDOMs = append(t.decodedDOMs, part)
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
	
	return idx, true, simpleString
}