package models

import "time"

type Options struct {
	RedisIp [4]byte // сделать поудобнее, одним полем
	RedisPort int // сделать поудобнее, одним полем

	// Количество ретраев (без учета первой попытки, то есть всего будет RetryAmount + 1 попыток).
	// Должно быть > 0. По умолчанию 3
	//
	// При всех неудачных ретраях возвращается ошибка ErrConnectionRetriesFailed
	RetryAmount int

	// Размер TCP-буфера отправки (в ядре), не меняется динамически!
	// По умолчанию действует лимит ОС - 208 КБ
	TCPSendBufLen int 
	// Размер TCP-буфера получения (в ядре), не меняется динамически!
	// По умолчанию действует лимит ОС - 208 КБ
	TCPReceiveBufLen int

	// ------------------------------------------------
	// Keep-alive
	//
	// Нужно ли включить Keep-alive
	SetKeepAlive bool
	// Кол-во секунд бездействия, чтобы началась проверка соединения
	KeepAliveIdle int
	// Кол-во секунд перед следующей проверкой
	KeepAliveInterval int
	// Максимальное кол-во проверок, прежде чем соединение будет закрыто
	KeepAliveCheckAmount int

	// ------------------------------------------------
	// Размер буфера отправки
	//
	// Минимальная граница, ниже которой размер не должен опускаться. Должен быть  <= SendBufMaxLen
	// По умолчанию 8 КБ
	SendBufMinLen int
	// Максимальная граница, выше которой размер не должен подниматься. Должен быть >= SendBufMinLen
	// По умолчанию 100 МБ
	SendBufMaxLen int
	// Средний размер, которого нужно придерживаться. По умолчанию равен SendBufMinLen
	// SendBufAvgLen int

	// ------------------------------------------------
	// Размеры буфера получения
	//
	// Минимальная граница, ниже которой размер не должен опускаться. Должен быть <= ReceiveBufMaxLen
	// По умолчанию 8 КБ
	ReceiveBufMinLen int
	// Максимальная граница, выше которой размер не должен подниматься. Должен быть >= ReceiveBufMinLen
	// По умолчанию 100 МБ
	ReceiveBufMaxLen int
	// Средний размер, которого нужно придерживаться. По умолчанию равен ReceiveBufMinLen
	// ReceiveBufAvgLen int

	// Таймаут поллинга
	// По умолчанию 100 миллисек
	//
	// При наступлении возвращается ошибка ErrPollTimeout
	PollingTimeout time.Duration
}

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