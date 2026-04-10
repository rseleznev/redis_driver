package translator

type Encoder interface {
	Encode([]byte, []any) error
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

func (t translator) Encode([]byte, []any) error
func (t translator) Decode([]byte) (any, error)