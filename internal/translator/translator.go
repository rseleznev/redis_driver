package translator

type Encoder interface {
	Encode([]byte, []any) ([]byte, error)
}

type Decoder interface {
	Decode([]byte) (any, error)
}

type translator struct {}

func NewEncoder() Encoder {
	return translator{}
}
func NewDecoder() Decoder {
	return translator{}
}

func (t translator) Encode([]byte, []any) ([]byte, error) {
	return nil, nil
}

func (t translator) Decode([]byte) (any, error) {
	return nil, nil
}