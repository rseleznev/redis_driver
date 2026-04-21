package models

import "time"

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

	// Размер буфера отправки
	//
	// Минимальная граница, ниже которой размер не должен опускаться. Должен быть  <= SendBufMaxLen
	// По умолчанию 8 КБ
	SendBufMinLen int
	// Максимальная граница, выше которой размер не должен подниматься. Должен быть >= SendBufMinLen
	// По умолчанию 100 МБ
	SendBufMaxLen int
	// Средний размер, которого нужно придерживаться. По умолчанию равен SendBufMinLen
	SendBufAvgLen int

	// Размеры буфера получения
	//
	// Минимальная граница, ниже которой размер не должен опускаться. Должен быть <= ReceiveBufMaxLen
	// По умолчанию 8 КБ
	ReceiveBufMinLen int
	// Максимальная граница, выше которой размер не должен подниматься. Должен быть >= ReceiveBufMinLen
	// По умолчанию 100 МБ
	ReceiveBufMaxLen int
	// Средний размер, которого нужно придерживаться. По умолчанию равен ReceiveBufMinLen
	ReceiveBufAvgLen int

	// Таймаут поллинга
	// При наступлении возвращается ошибка ErrPollTimeout
	PollingTimeout time.Duration
}

type DOMPart struct {
	PartType string // тип элемента

	ValueLenBytes []byte // длина значения в виде нескольких байт (на случай частичного декодирования)
	ValueLen int // длина значения в байтах в виде одного числа (для простых типов)
	Value []byte // значение (для простых типов)

	ContentLenBytes []byte // длина контента в виде нескольких байт (на случай частичного декодирования)
	ContentLen int // кол-во элементов внутри одним числом (для составных типов)
	Content []DOMPart // дочерние элементы (для составных типов)
	
	TotalBytesLen int // кол-во байтов (для расчета емкости итогового среза) можно удалить
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

type SendBuf struct {
	WritePos int
	SentBytes int
	Buf []byte
}

type RecvBuf struct {
	WritePos int
	Buf []byte
}