package models

import (
	"syscall"
)

type EpollEvent struct {
	Type string
	Event syscall.EpollEvent
}