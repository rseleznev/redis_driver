package translator

import "github.com/rseleznev/redis_driver/internal/models"

type Translator struct {
	sendBuf *models.SendBuf

	decodingData []byte
}

func NewTranslator() *Translator {
	return &Translator{}
}

func (t *Translator) setSendBuf(buf *models.SendBuf) {
	t.sendBuf = buf
}

func (t *Translator) sendBufLen() int {
	return len(t.sendBuf.Buf)
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