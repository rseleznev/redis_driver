package translator

import (
	"github.com/rseleznev/redis_driver/internal/models"
)

type Translator struct {
	// интерфейс, который преобразует корневой DOM-объект (со всем содержимым) в формат RESP
	serializator
}

type serializator interface {
	serializeDOMToRESP([]byte, models.DOMPart) []byte
}

func NewTranslator() Translator {
	return Translator{
		serializator: serializer{},
	}
}

func (t Translator) Decode([]byte) (any, error) {
	return nil, nil
}