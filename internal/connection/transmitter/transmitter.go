package transmitter

type Transmitter struct {}

func NewSender() Transmitter {
	return Transmitter{}
}

func NewReceiver() Transmitter {
	return Transmitter{}
}

func (s Transmitter) Send(data []byte) (int, error) {
	
	
	return 0, nil
}

func (s Transmitter) Receive(data []byte) error {
	
	
	return nil
}