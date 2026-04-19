package translator

import "github.com/rseleznev/redis_driver/internal/models"

type Translator struct {
	// декодирование не завершено
	decodeTrunc bool

	// незавершенный DOM-объект
	domTrunc models.DOMPart
}

func NewTranslator() *Translator {
	return &Translator{}
}