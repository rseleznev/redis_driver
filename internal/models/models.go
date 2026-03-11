package models

type Command struct {
	Operation string
	SendingData []byte
	ResultChan chan []byte
}

type DOMPart struct {
	PartType string // тип элемента

	ContentLen int // кол-во элементов внутри (для составных типов)

	ValueLen int // длина значения (для простых типов)
	Value []byte // значение (для простых типов)
}