package models

import (
	"errors"
)

var (
	// epoll
	ErrSocketEvent = errors.New("redis_driver: error event has happened on socket") // EPOLLERR event
	ErrSocketHUPEvent = errors.New("redis_driver: HUP error event has happened on socket") // EPOLLHUP event
	ErrSocketRDHUPEvent = errors.New("redis_driver: RDHUP error event has happened on socket") // EPOLLRDHUP event
	ErrEpollNoMemory = errors.New("redis_driver: not enought memory available to create an epoll instance") // ENOMEM
	ErrEpollBadFD = errors.New("redis_driver: epoll file descriptor or socket file descriptor is not a valid file descriptor") // EBADF
	ErrSocketAlreadyAdded = errors.New("redis_driver: socket already in interest list") // EEXIST
	ErrSocketNotAdded = errors.New("redis_driver: socket not added in interest list") // ENOENT

	// socket
	ErrSocketNoAccess = errors.New("redis_driver: no access to socket") // EACCES
	ErrSocketNoMemory = errors.New("redis_driver: not enought memory available to create a socket") // ENOBUFS or ENOMEM
	ErrSocketLocalPortInUse = errors.New("redis_driver: local port already in use")  // EADDRINUSE
	ErrSocketNoLocalPorts = errors.New("redis_driver: not enought free local ports") // EADDRNOTAVAIL
	ErrAddrBadParams = errors.New("redis_driver: bad address given") // EAFNOSUPPORT
	ErrSocketBadFD = errors.New("redis_driver: socket file descriptor is not a valid file descriptor") // EBADF, ENOTSOCK

	// connection
	ErrConnectionInProcess = errors.New("redis_driver: connection attempt already in process") // EALREADY
	ErrConnectionRetriesFailed = errors.New("redis_driver: all connection retries failed")
	ErrConnectionRefused = errors.New("redis_driver: connection is refused") // ECONNREFUSED
	ErrServerUnreachable = errors.New("redis_driver: server is unreachable") // EHOSTUNREACH
	ErrNetUnreachable = errors.New("redis_driver: network is unreachable") // ENETUNREACH
	ErrConnectionReset = errors.New("redis_driver: connection is reset by server") // ECONNRESET
	ErrNotConnected = errors.New("redis_driver: socket is not connected") // ENOTCONN
	ErrConnectionClosed = errors.New("redis_driver: connection is closed by server") // EPIPE

	// common
	ErrTooManyFilesInProcess = errors.New("redis_driver: per-process limit of open file descriptors has been reached") // EMFILE
	ErrTooManyFilesInSystem = errors.New("redis_driver: system-wide limit of open file descriptors has been reached") // ENFILE
	ErrSignalInterruption = errors.New("redis_driver: operation is interrupted by signal") // EINTR
	ErrNoMemory = errors.New("redis_driver: not enought memory for requested operation") // ENOMEM
	ErrSpaceAddress = errors.New("redis_driver: an invalid user space address") // EFAULT
	ErrBadValue = errors.New("redis_driver: invalid argument passed") // EINVAL

	// sending
	ErrSendMsgTrunc = errors.New("redis_driver: sended message truncated")
	ErrSendNoAccess = errors.New("redis_driver: no permission for sending") // EACCES

	// receiving
	ErrRecvMsgTrunc = errors.New("redis_driver: received message truncated")
	ErrRecvMsgCTrunc = errors.New("redis_driver: received oob-message truncated")
)