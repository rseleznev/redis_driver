package translator

import (
	"github.com/rseleznev/redis_driver/internal/models"
)

type Translator struct {
	// интерфейс, который строит один DOM-объект
	builder

	// интерфейс, который преобразует корневой DOM-объект (со всем содержимым) в формат RESP
	serializator
}

type builder interface {
	buildDOMPart(any) (models.DOMPart, error)
}

type serializator interface {
	serializeDOMToRESP([]byte, models.DOMPart) []byte
}

func NewTranslator() Translator {
	return Translator{
		builder: domBuilder{},
		serializator: domSerializator{},
	}
}

func (t Translator) Decode([]byte) (any, error) {
	return nil, nil
}