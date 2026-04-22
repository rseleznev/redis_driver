package translator

import (
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
			return err
		}
	}
	
	return nil
}

func (t *Translator) writeArr(paramsLen int) error {
	paramsLenStr := strconv.Itoa(paramsLen)
	paramsLenBytes := []byte(paramsLenStr)
	
	arrLen := 1 + len(paramsLenBytes) + 2

	if arrLen <= t.sendBufLen() {
		return models.ErrSendBufTooShort
	}

	t.writeByte('*')

	for _, v := range paramsLenBytes {
		t.writeByte(v)
	}

	t.writeByte('\r')
	t.writeByte('\n')

	return nil
}

func (t *Translator) writeString(cmdPart any) error {
	valueBytes, ok := cmdPart.([]byte)
	if !ok {
		return models.ErrUnsupportedDataType
	}
	valueLenStr := strconv.Itoa(len(valueBytes))
	valueLenBytes := []byte(valueLenStr)
	valueLen := 1 + len(valueLenBytes) + len(valueBytes) + 4

	if valueLen <= t.sendBufLen() {
		return models.ErrSendBufTooShort
	}

	t.writeByte('$')

	for _, v := range valueLenBytes {
		t.writeByte(v)
	}

	t.writeByte('\r')
	t.writeByte('\n')

	for _, v := range valueBytes {
		t.writeByte(v)
	}

	t.writeByte('\r')
	t.writeByte('\n')

	return nil
}

func (t *Translator) writeByte(b byte) {
	t.sendBuf.Buf[t.sendBuf.WritePos] = b
	t.sendBuf.WritePos++
}