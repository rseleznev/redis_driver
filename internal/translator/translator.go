package translator

type Translator struct {
	decodingData []byte
}

func NewTranslator() *Translator {
	return &Translator{}
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