package transmitter

import "github.com/rseleznev/redis_driver/internal/models"

type Transmitter struct {
	SocketFd int
}

func NewTransmitter(socketFd int) Transmitter {
	return Transmitter{
		SocketFd: socketFd,
	}
}

func (s Transmitter) Send(data []byte) (int, error) {
	
	
	return 0, nil
}

func (s Transmitter) Receive(data *models.RecvBuf) error {
	
	return nil
}