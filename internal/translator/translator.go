package translator

import "github.com/rseleznev/redis_driver/internal/models"

type Translator struct {
	decodingData []byte
	// продолжаем декодирование или декодируем сначала
	decodeProceeding bool

	// незаконченный декодируемый объект
	decodingDOMPart *models.DOMPart

	// декодированный DOM-объект
	decodedDOM *models.DOMPart
}

func NewTranslator() *Translator {
	return &Translator{}
}

func (t *Translator) isDecodeProceeding() bool {
	return t.decodeProceeding
}

func (t *Translator) setDecodeProceeding() {
	t.decodeProceeding = true
}

func (t *Translator) setDecodingData(d []byte) {
	t.decodingData = d
}

func (t *Translator) decodingDataLen() int {
	return len(t.decodingData)
}

func (t *Translator) isDataEnded(idx int) bool {
	return idx >= t.decodingDataLen()
}