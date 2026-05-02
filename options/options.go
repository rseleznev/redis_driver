package options

import (
	"time"
)

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
	// По умолчанию 300 секунд (5 минут)
	KeepAliveIdle int
	// Кол-во секунд перед следующей проверкой
	// По умолчанию 60 секунд
	KeepAliveInterval int
	// Максимальное кол-во проверок, прежде чем соединение будет закрыто
	// По умолчанию 5
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
	// По умолчанию 50 миллисек
	//
	// При наступлении возвращается ошибка ErrPollTimeout
	PollingTimeout time.Duration
}

// InitOptions инициализирует параметры соединения
func InitOptions(opts *Options) {
	if opts.RetryAmount <= 0 {
		opts.RetryAmount = 3
	}
	opts.RetryAmount += 1 // + первая попытка, которая не является ретраем

	if opts.SetKeepAlive {
		if opts.KeepAliveIdle <= 0 {
			opts.KeepAliveIdle = 300
		}
		if opts.KeepAliveInterval <= 0 {
			opts.KeepAliveInterval = 60
		}
		if opts.KeepAliveCheckAmount <= 0 {
			opts.KeepAliveCheckAmount = 5
		}
	}

	// буфер отправки
	if opts.SendBufMinLen <= 0 {
		opts.SendBufMinLen = 8 * 1024
	}
	if opts.SendBufMaxLen <= 0 {
		opts.SendBufMaxLen = 100 * 1024 * 1024
	}
	if opts.SendBufMinLen > opts.SendBufMaxLen {
		opts.SendBufMinLen = 8 * 1024
		opts.SendBufMaxLen = 100 * 1024 * 1024
	}

	// буфер получения
	if opts.ReceiveBufMinLen <= 0 {
		opts.ReceiveBufMinLen = 8 * 1024
	}
	if opts.ReceiveBufMaxLen <= 0 {
		opts.ReceiveBufMaxLen = 100 * 1024 * 1024
	}
	if opts.ReceiveBufMinLen > opts.ReceiveBufMaxLen {
		opts.ReceiveBufMinLen = 8 * 1024
		opts.ReceiveBufMaxLen = 100 * 1024 * 1024
	}


	if opts.PollingTimeout <= 0 {
		opts.PollingTimeout = time.Millisecond*50
	}
}