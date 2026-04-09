package connection

import "github.com/rseleznev/redis_driver/internal/models"

type Connector interface {

}

func NewC(opts models.Options) (Connector, error)