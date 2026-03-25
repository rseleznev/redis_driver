package message

import (
	"fmt"
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

// SerializeCommand сериализует любую команду с параметрами в формат RESP
func SerializeCommand(params ...any) ([]byte, error) {
	var bytesLen int
	paramsAmount := len(params)

	parts := make([]models.DOMPart, paramsAmount)
	var err error

	// Переводим все переданные параметры в DOM
	for i, v := range params {
		parts[i], err = buildDOMPart(v)
		if err != nil {
			return nil, err
		}
		bytesLen += parts[i].TotalBytesLen
	}

	// Корневой массив (команда - всегда массив)
	arr := models.DOMPart{
		PartType: "array",

		ContentLen: len(parts),
		Content: parts,

		TotalBytesLen: bytesLen + (len(parts)*2), // кол-во байтов всего контента + '\r' и '\n' * кол-во элементов (не точно!)
	}
	
	// Сериализуем в RESP
	result := serializeDOMToRESP(arr)
	
	return result, nil
}

// buildDOMPart переводит тип данных Go в элемент DOM. Поддерживаются только: 
// string, 
// []byte, 
func buildDOMPart(input any) (models.DOMPart, error) {
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

	default:
		return models.DOMPart{}, models.ErrWrongDataType
		
	}

	return part, nil
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