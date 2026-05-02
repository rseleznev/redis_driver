package redis_driver

import (
	"time"

	"github.com/rseleznev/redis_driver/internal/models"
)

// initOptions инициализирует параметры соединения
func initOptions(opts *models.Options) {
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