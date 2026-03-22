package models

import (
	"errors"
)

var (
	// epoll
	ErrSocketEvent = errors.New("redis_driver: error event has happened on socket")
	ErrSocketHUPEvent = errors.New("redis_driver: HUP error event has happened on socket")
	ErrSocketRDHUPEvent = errors.New("redis_driver: RDHUP error event has happened on socket")

	// socket
	ErrSocketNoAccess = errors.New("redis_driver: no access to socket") // EACCES
	ErrTooManyFilesInProcess = errors.New("redis_driver: per-process limit of open file descriptors has been reached") // EMFILE
	ErrTooManyFilesInSystem = errors.New("redis_driver: system-wide limit of open file descriptors has been reached") // ENFILE
	ErrSocketNoMemory = errors.New("redis_driver: not enought memory available") // ENOBUFS or ENOMEM
	ErrSocketLocalPortInUse = errors.New("redis_driver: local port already in use")  // EADDRINUSE
	ErrSocketNoLocalPorts = errors.New("redis_driver: not enought free local ports") // EADDRNOTAVAIL
	ErrAddrBadParams = errors.New("redis_driver: bad address given") // EAFNOSUPPORT
	ErrConnectionInProcess = errors.New("redis_driver: connection attempt already in process") // EALREADY
	ErrSocketBadFD = errors.New("redis_driver: socket file descriptor is not a valid descriptor") // EBADF
	ErrSignalInterruption = errors.New("redis_driver: operation is interrupted by signal") // EINTR

	// connection
	ErrConnectionRetriesFailed = errors.New("redis_driver: all connection retries failed")
	ErrConnectionRefused = errors.New("redis_driver: connection is refused") // ECONNREFUSED
	ErrServerUnreachable = errors.New("redis_driver: server is unreachable") // EHOSTUNREACH
	ErrNetUnreachable = errors.New("redis_driver: network is unreachable") // ENETUNREACH
)