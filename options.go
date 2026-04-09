package redis_driver

import "github.com/rseleznev/redis_driver/internal/models"

// initOptions инициализирует параметры соединения
func initOptions(opts *models.Options) {
	if opts.RetryAmount == 0 {
		opts.RetryAmount = 3
	}

	if opts.ReceiveBufMinLen == 0 {
		opts.ReceiveBufMinLen = 8 * 1024
	}
	if opts.ReceiveBufMaxLen == 0 {
		opts.ReceiveBufMaxLen = 100 * 1024 * 1024
	}
	if opts.ReceiveBufMinLen > opts.ReceiveBufMaxLen {
		opts.ReceiveBufMinLen = 8 * 1024
		opts.ReceiveBufMaxLen = 100 * 1024 * 1024
	}
	if opts.ReceiveBufAvgLen == 0 {
		opts.ReceiveBufAvgLen = opts.ReceiveBufMinLen
	}
}
