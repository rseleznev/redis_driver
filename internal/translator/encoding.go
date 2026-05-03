package translator

import (
	"math"
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

func (t *Translator) Encode(buf *models.SendBuf, params []any) error {
	t.setSendBuf(buf)

	err := t.writeArr(len(params))
	if err != nil {
		return err
	}

	for _, v := range params {
		err = t.writeString(v)
		if err != nil {
			t.resetWritePos()
			return err
		}
	}
	
	return nil
}

func (t *Translator) writeArr(paramsLen int) error {
	paramsLenStr := strconv.Itoa(paramsLen)
	// paramsLenBytes := []byte(paramsLenStr)
	paramsLenInBytes := int(math.Log10(float64(paramsLen)))+1
	
	arrLen := 1 + paramsLenInBytes + 2

	if arrLen >= t.sendBufLen() {
		return models.ErrSendBufTooShort
	}

	t.writeByte('*')

	// for _, v := range paramsLenBytes {
	// 	t.writeByte(v)
	// }

	for i:=0; i<len(paramsLenStr); i++ {
		t.writeByte(paramsLenStr[i])
	}

	t.writeByte('\r')
	t.writeByte('\n')

	return nil
}

func (t *Translator) writeString(cmdPart any) error {
	var valueBytes []byte
	
	switch cmdPart := cmdPart.(type) {
	case string:
		valueBytes = []byte(cmdPart)

	case []byte:
		valueBytes = cmdPart

	default:
		return models.ErrUnsupportedDataType

	}

	valueLenStr := strconv.Itoa(len(valueBytes))
	valueLenBytes := []byte(valueLenStr)
	valueLen := 1 + len(valueLenBytes) + len(valueBytes) + 4

	if valueLen > t.sendBufLen() {
		return models.ErrSendBufTooShort
	}

	t.writeByte('$')

	// for _, v := range valueLenBytes {
	// 	t.writeByte(v)
	// }

	for i:=0; i<len(valueLenBytes); i++ {
		t.writeByte(valueLenBytes[i])
	}

	t.writeByte('\r')
	t.writeByte('\n')

	// for _, v := range valueBytes {
	// 	t.writeByte(v)
	// }

	for i:=0; i<len(valueBytes); i++ {
		t.writeByte(valueBytes[i])
	}

	t.writeByte('\r')
	t.writeByte('\n')

	return nil
}

func (t *Translator) writeByte(b byte) {
	t.sendBuf.Buf[t.sendBuf.WritePos] = b
	t.sendBuf.WritePos++
}

func (t *Translator) resetWritePos() {
	t.sendBuf.WritePos = 0
}