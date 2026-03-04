package models

type Command struct {
	Operation string
	Key []byte
	ResultChan chan []byte
}