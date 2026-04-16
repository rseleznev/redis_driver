package translator

import (
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

func (t Translator) Encode(buf *models.SendBuf, params []any) error {
	paramsAmount := len(params)

	parts := make([]models.DOMPart, paramsAmount)
	var err error

	// Переводим все переданные параметры в DOM
	for i, v := range params {
		parts[i], err = t.buildDOMPart(v)
		if err != nil {
			return err
		}
	}

	// Корневой массив (команда - всегда массив)
	arr := models.DOMPart{
		PartType: "array",

		ContentLen: len(parts),
		Content: parts,
	}
	
	// Сериализуем в RESP
	buf.Buf = t.serializeDOMToRESP(buf.Buf, arr)

	// надо сделать замороченную логику с указанием, до какого индекса буфера мы записали
	
	return nil
}

// buildDOMPart переводит тип данных Go в элемент DOM. Поддерживаются только: 
// string, 
// []byte, 
func (t Translator) buildDOMPart(input any) (models.DOMPart, error) {
	var part models.DOMPart

	switch input := input.(type) {
	case string:		
		part.PartType = "string"
		part.ValueLen = len(input)
		part.Value = []byte(input)

	case []byte:
		part.PartType = "string"
		part.ValueLen = len(input)
		part.Value = input

	default:
		return models.DOMPart{}, models.ErrUnsupportedDataType
		
	}

	return part, nil
}

// serializeDOMToRESP сериализует корневой DOM в RESP. Предполагается, что на вход поступит корневой DOM-элемент,
// который содержит весь контент внутри себя
func (t Translator) serializeDOMToRESP(buf []byte, input models.DOMPart) []byte {
	if input.PartType != "array" {
		panic("некорректный корневой элемент")
	}

	arrPart := t.serializeDOMPartToRESP(input)
	buf = append(buf, arrPart...)

	for _, v := range input.Content {
		part := t.serializeDOMPartToRESP(v)
		buf = append(buf, part...)
	}

	return buf
}

// serializeDOMPartToRESP сериализует отдельный DOM элемент в формат RESP
func (t Translator) serializeDOMPartToRESP(input models.DOMPart) []byte {

	switch input.PartType {
	case "string":
		result := make([]byte, 0, input.ValueLen + 6) // не точный расчет
		result = append(result, '$')

		vls := strconv.Itoa(input.ValueLen)
		vlb := []byte(vls)

		for _, v := range vlb {
			result = append(result, v)
		}
		result = append(result, '\r', '\n')
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
		return nil
	}
}