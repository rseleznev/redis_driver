package models

type Options struct {
	RedisIp [4]byte // сделать поудобнее, одним полем
	RedisPort int // сделать поудобнее, одним полем
	RetryAmount int // количество ретраев
	// таймауты
	// размер буфера отправки
	// размер буфера получения
}

type Command struct {
	Operation string
	SendingData []byte
	ResultChan chan []byte
	ErrChan chan error
}

type DOMPart struct {
	PartType string // тип элемента

	ValueLen int // длина значения в байтах (для простых типов)
	Value []byte // значение (для простых типов)

	ContentLen int // кол-во элементов внутри (для составных типов)
	Content []DOMPart // дочерние элементы (для составных типов)
	
	TotalBytesLen int // кол-во байтов (для расчета емкости итогового среза)
}