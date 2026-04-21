package translator

import (
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

func (t *Translator) DecodeWithProceeding(input []byte) {
	// закидываем срез в структуру
	t.setDecodingData(input)

	var idx int
	var finished bool
	var part models.DOMPart

	// декодируем с начала или продолжаем
	if t.isDecodeProceeding() {
		// декодирование продолжается не с начала
		// дособираем незаконченный объект
		idx = t.makePartDone()
		idx++
	}
	if t.isDataEnded(idx) {
		return
	}

	// декодируем с начала
	for {
		idx, finished, part = t.parsePartNew(idx)
		if !finished {
			t.decodingDOMPart = &part

			break
		}
		t.decodedDOM = append(t.decodedDOM, part)
		
		idx++
		if idx >= t.decodingDataLen() {
			break
		}
	}

	// указываем, что нужно будет продолжить декодирование
	t.setDecodeProceeding()
}

func (t *Translator) parsePartNew(idx int) (int, bool, models.DOMPart) {
	var finished bool
	var part models.DOMPart

	// определяем тип объекта
	switch t.decodingData[idx] {
	case '+': // Simple string 
		idx, finished, part = t.parseSimpleString(idx)

	case '$': // Bulk strings
		idx, finished, part = t.parseBulkString(idx)
	
	case '%': // Maps
		idx, finished, part = t.parseMap(idx)

	case ':': // Integers
		idx, finished, part = t.parseInteger(idx)
	
	case '*': // Arrays
		idx, finished, part = t.parseArray(idx)

	case '_': // Nil
		// парсим nil

	case '-': // Simple Errors
		// парсим ошибку

	default:
		panic("unsupported RESP3 data type")
	}

	return idx, finished, part
}

func (t *Translator) parseSimpleString(idx int) (int, bool, models.DOMPart) {
	var simpleString models.DOMPart

	simpleString.PartType = "s_string"
	idx++

	for {
		if t.isDataEnded(idx) {
			return idx, false, simpleString
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnded(idx) {
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

	if t.isDataEnded(idx) {
		return idx, false, bulkString
	}
	idx, finished, bulkString.ValueLenBytes = t.parsePartLenNew(idx)
	
	if !finished {
		return idx, false, bulkString
	}
	lenString := string(bulkString.ValueLenBytes)
	bulkString.ValueLen, _ = strconv.Atoi(lenString)
	idx++

	for {
		if t.isDataEnded(idx) {
			return idx, false, bulkString
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnded(idx) {
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

func (t *Translator) parseMap(idx int) (int, bool, models.DOMPart) {
	var m models.DOMPart
	var finished bool

	m.PartType = "map"
	idx++

	if t.isDataEnded(idx) {
		return idx, false, m
	}
	idx, finished, m.ContentLenBytes = t.parsePartLenNew(idx)

	if !finished {
		return idx, false, m
	}
	lenString := string(m.ContentLenBytes)
	m.ContentLen, _ = strconv.Atoi(lenString)

	var mapPart models.DOMPart
	partsToCollect := m.ContentLen*2

	for partsToCollect > 0 {
		idx++
		if t.isDataEnded(idx) {
			return idx, false, m
		}

		idx, finished, mapPart = t.parsePartNew(idx)
		m.Content = append(m.Content, mapPart)

		if !finished {
			return idx, false, m
		}
		partsToCollect--
	}

	return idx, true, m
}

func (t *Translator) parseInteger(idx int) (int, bool, models.DOMPart) {
	var integer models.DOMPart
	
	integer.PartType = "int"
	idx++

	for {
		if t.isDataEnded(idx) {
			return idx, false, integer
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnded(idx) {
				return idx, false, integer
			}

			if t.decodingData[idx] == '\n' {
				break
			}
		}

		integer.Value = append(integer.Value, t.decodingData[idx])
		idx++
	}

	return idx, true, integer
}

func (t *Translator) parseArray(idx int) (int, bool, models.DOMPart) {
	var arr models.DOMPart
	var finished bool

	arr.PartType = "array"
	idx++

	if t.isDataEnded(idx) {
		return idx, false, arr
	}

	idx, finished, arr.ContentLenBytes = t.parsePartLenNew(idx)
	if !finished {
		return idx, false, arr
	}
	lenString := string(arr.ContentLenBytes)
	arr.ContentLen, _ = strconv.Atoi(lenString)

	var arrPart models.DOMPart
	partsToCollect := arr.ContentLen

	for partsToCollect > 0 {
		idx++
		if t.isDataEnded(idx) {
			return idx, false, arr
		}

		idx, finished, arrPart = t.parsePartNew(idx)
		arr.Content = append(arr.Content, arrPart)

		if !finished {
			return idx, false, arr
		}
		partsToCollect--
	}

	return idx, true, arr
}

func (t *Translator) parsePartLenNew(idx int) (int, bool, []byte) {
	valueLenBytes := make([]byte, 0, 5)
	idx++
	
	for {
		if t.isDataEnded(idx) {
			return idx, false, valueLenBytes
		}
		
		if t.decodingData[idx] == '\r' {
			idx++
			if t.isDataEnded(idx) {
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

func (t *Translator) makePartDone() int {
	var idx int
	finished := true

	decodingPart := t.decodingDOMPart

	switch decodingPart.PartType {
	case "s_string":
		for {
			if t.isDataEnded(idx) {
				finished = false
				break
			}
			
			if t.decodingData[idx] == '\r' {
				idx++
				if t.isDataEnded(idx) {
					finished = false
					break
				}

				if t.decodingData[idx] == '\n' {
					break
				}
			}

			decodingPart.Value = append(decodingPart.Value, t.decodingData[idx])
			idx++
		}
		if !finished {
			return idx
		}
		
		t.decodingDOMPart = nil
		t.decodedDOM = append(t.decodedDOM, *decodingPart)
	}

	return idx
}