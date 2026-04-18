package transmitter

import "github.com/rseleznev/redis_driver/internal/models"

type Transmitter struct {
	SocketFd int
}

func NewTransmitter(socketFd int) *Transmitter {
	return &Transmitter{
		SocketFd: socketFd,
	}
}

func (t *Transmitter) Send(data []byte) (int, error) {
	
	
	return 0, nil
}

func (t *Transmitter) Receive(data *models.RecvBuf) error {
	
	return nil
}

func (t *Transmitter) ChangeSocketFd(newSocketFd int) {
	t.SocketFd = newSocketFd
}