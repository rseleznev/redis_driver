package translator

import (
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

func (t Translator) Encode(buf *models.SendBuf, params []any) error {
	paramsLen := len(params)
	paramsLenStr := strconv.Itoa(paramsLen)
	paramsLenBytes := []byte(paramsLenStr)
	
	
	
	return nil
}