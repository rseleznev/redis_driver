package message

import (
	"fmt"
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

// SerializeGetCommand сериализует команду GET в формат RESP
func SerializeGetCommand(key string) []byte {
	var bytesLen int
	
	// Добавляем команду
	c := buildDOMPart("GET")
	bytesLen += c.TotalBytesLen

	// Добавляем ключ
	k := buildDOMPart(key)
	bytesLen += k.TotalBytesLen

	parts := make([]models.DOMPart, 2)
	parts[0] = c
	parts[1] = k

	// Корневой массив
	arr := models.DOMPart{
		PartType: "array",

		ContentLen: len(parts),
		Content: parts,

		TotalBytesLen: bytesLen + (len(parts)*2), // кол-во байтов всего контента + '\r' и '\n' * кол-во элементов (не точно!)
	}

	result := serializeDOMToRESP(arr)

	return result
}

// SerializeSetCommand сериализует команду SET в формат RESP
func SerializeSetCommand(key string, value any, dur int) []byte {
	var bytesLen int

	// Добавляем команду
	c := buildDOMPart("SET")
	bytesLen += c.TotalBytesLen

	// Добавляем ключ
	k := buildDOMPart(key)
	bytesLen += k.TotalBytesLen

	// Добавляем значение
	v := buildDOMPart(value)
	bytesLen += v.TotalBytesLen

	// Добавляем длительность
	ex := buildDOMPart("EX")
	bytesLen += ex.TotalBytesLen

	durString := strconv.Itoa(dur)
	dr := buildDOMPart(durString)
	bytesLen += dr.TotalBytesLen

	parts := make([]models.DOMPart, 5)
	parts[0] = c
	parts[1] = k
	parts[2] = v
	parts[3] = ex // не закидывать, если не указано
	parts[4] = dr // не закидывать, если не указано

	// Корневой массив (команда - всегда массив)
	arr := models.DOMPart{
		PartType: "array",

		ContentLen: len(parts),
		Content: parts,

		TotalBytesLen: bytesLen + (len(parts)*2), // кол-во байтов всего контента + '\r' и '\n' * кол-во элементов (не точно!)
	}

	// Сериализуем в RESP
	result := serializeDOMToRESP(arr)
	
	return result
}

// buildDOMPart переводит тип данных Go в элемент DOM
// Поддерживаются только: 
// string, 
// []byte, 
// map[string]string
func buildDOMPart(input any) models.DOMPart {
	var part models.DOMPart

	switch input := input.(type) {
	case string:		
		part.PartType = "string"
		part.ValueLen = len(input)
		part.Value = []byte(input)
		part.TotalBytesLen = len(input)

	case []byte:
		part.PartType = "string"
		part.ValueLen = len(input)
		part.Value = input
		part.TotalBytesLen = len(input)

	case map[string]string:
		part.PartType = "map"
		part.ContentLen = len(input)

		cntnt := make([]models.DOMPart, 0, part.ContentLen*2)
		part.Content = cntnt

		var bytesLen int

		for k, v := range input {
			var keyPart, valuePart models.DOMPart

			// Обрабатываем ключ
			keyPart.PartType = "string"
			keyPart.ValueLen = len(k)
			keyPart.Value = []byte(k)
			keyPart.TotalBytesLen = len(k)

			bytesLen += keyPart.TotalBytesLen
			part.Content = append(part.Content, keyPart)

			// Обрабатываем значение
			valuePart.PartType = "string"
			valuePart.ValueLen = len(v)
			valuePart.Value = []byte(v)
			valuePart.TotalBytesLen = len(v)

			bytesLen += valuePart.TotalBytesLen
			part.Content = append(part.Content, valuePart)
		}
		part.TotalBytesLen = bytesLen

	default:
		panic("redis_driver: неподдерживаемый тип данных")
	}

	return part
}

// serializeDOMToRESP сериализует корневой DOM в RESP. Предполагается, что на вход поступит корневой DOM-элемент,
// который содержит весь контент внутри себя
func serializeDOMToRESP(input models.DOMPart) []byte {
	result := make([]byte, 0, input.TotalBytesLen)
	
	if input.PartType != "array" {
		panic("некорректный корневой элемент")
	}

	arrPart := serializeDOMPartToRESP(input)
	result = append(result, arrPart...)

	for _, v := range input.Content {
		part := serializeDOMPartToRESP(v)
		result = append(result, part...)

		if v.PartType == "map" {
			for _, mv := range v.Content {
				mapPart := serializeDOMPartToRESP(mv)
				result = append(result, mapPart...)
			}
		}
	}

	return result
}

// serializeDOMPartToRESP сериализует отдельный DOM элемент в формат RESP
func serializeDOMPartToRESP(input models.DOMPart) []byte {

	switch input.PartType {
	case "string":
		result := make([]byte, 0, input.ValueLen + 6) // не точный расчет
		result = append(result, '$')

		vl := input.ValueLen
		if vl > 99 {
			panic("слишком длинная строка!") // заглушка
		}
		if vl > 9 {
			fstDg := vl / 10
			scndDg := vl % 10
			fstDgS := strconv.Itoa(fstDg)
			scndDgS := strconv.Itoa(scndDg)
			result = append(result, fstDgS[0], scndDgS[0], '\r', '\n')
		} else {
			vls := strconv.Itoa(vl)
			result = append(result, vls[0], '\r', '\n')
		}
		result = append(result, input.Value...)
		result = append(result, '\r', '\n')

		return result

	case "map":
		result := make([]byte, 0, 4)
		result = append(result, '%')

		vl := input.ContentLen
		if vl > 99 {
			panic("слишком длинная строка!") // заглушка
		}
		if vl > 9 {
			fstDg := vl / 10
			scndDg := vl % 10
			fstDgS := strconv.Itoa(fstDg)
			scndDgS := strconv.Itoa(scndDg)
			result = append(result, fstDgS[0], scndDgS[0], '\r', '\n')
		} else {
			vls := strconv.Itoa(vl)
			result = append(result, vls[0], '\r', '\n')
		}

		return result

	case "array":
		result := make([]byte, 4)

		ls := strconv.Itoa(input.ContentLen)
		lb := []byte(ls)

		result[0] = '*'
		result[1] = lb[0]
		result[2] = '\r'
		result[3] = '\n'

		return result

	default:
		fmt.Println("Сериализация неподдерживаемого типа: ", input.PartType)
		return nil
	}
}