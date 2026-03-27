package redis_driver

import "github.com/rseleznev/redis_driver/internal/models"

func initOptions(opts *models.Options) {
	if opts.RetryAmount == 0 {
		opts.RetryAmount = 3
	}

	
}