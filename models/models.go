package models

type Command struct {
	Operation string
	SendingData []byte
	ResultChan chan []byte
}