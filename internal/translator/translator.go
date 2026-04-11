package translator

type Translator struct {}

func NewTranslator() Translator {
	return Translator{}
}

func (t Translator) Encode([]byte, []any) ([]byte, error) {
	return nil, nil
}

func (t Translator) Decode([]byte) (any, error) {
	return nil, nil
}