package models

type Options struct {
	RedisIp [4]byte // сделать поудобнее, одним полем
	RedisPort int // сделать поудобнее, одним полем

	// Количество ретраев. По умолчанию 3
	RetryAmount int 

	// Размер TCP-буфера отправки (в ядре), не меняется динамически!
	// По умолчанию действует лимит ОС - 208 КБ
	TCPSendBufLen int 
	// Размер TCP-буфера получения (в ядре), не меняется динамически!
	// По умолчанию действует лимит ОС - 208 КБ
	TCPReceiveBufLen int

	// таймаут

	// размер буфера отправки

	// Размеры буфера получения
	//
	// Минимальная граница, ниже которой размер никогда не опустится. Должен быть <= ReceiveBufMaxLen
	// По умолчанию 8 КБ
	ReceiveBufMinLen int
	// Максимальная граница, выше которой размер никогда не поднимется. Должен быть >= ReceiveBufMinLen
	// По умолчанию 100 МБ
	ReceiveBufMaxLen int
	// Средний размер, которого нужно придерживаться. По умолчанию равен ReceiveBufMinLen
	ReceiveBufAvgLen int
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

// Подумать над синхронизацией
type PollingUnit struct {
	SocketFd int
	EventType string // connect, income, outcome
	ResultChan chan error // канал, чтобы вызывающий поток заблокировался на чтении
}

type PollingResult struct {
	Err error
}