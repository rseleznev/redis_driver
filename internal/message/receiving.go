package message

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
	"github.com/rseleznev/redis_driver/internal/models"
)

// Receive получает сообщение по указанному socket
func Receive(socketFd, retriesAvailable int, recvBuf []byte) ([]byte, error) {
	var result []byte
	var err error
	attempt := 1 // счетчик ретраев

	for {
		result, err = tryReceive(socketFd, recvBuf)
		if err != nil {
			if errors.Is(err, syscall.EWOULDBLOCK) {
				epoll.Wait()
				continue
			}
			
			// Проверяем ошибки, при которых нет смысла делать ретраи
			switch err {
			case models.ErrSocketBadFD, models.ErrConnectionRefused, models.ErrSpaceAddress, models.ErrSignalInterruption, 
			models.ErrBadValue, models.ErrNoMemory, models.ErrNotConnected, models.ErrConnectionClosed:
				return nil, err

			}

			// В буфер получения влезло не все
			if err == models.ErrRecvMsgTrunc {
				// нужно сложить прочитанное в отдельный буфер и идти за оставшимися данными
			}

			// делаем ретраи
			if attempt < retriesAvailable {
				attempt++
				continue
			} else {
				return nil, models.ErrConnectionRetriesFailed
			}
		}
		break
	}
	
	// Проверки события
	err = epoll.ProcessEvent(socketFd)
	if err != nil {
		return nil, err
	}

	return result, err
}

// tryReceive делает одну попытку прочитать ответ сервера
func tryReceive(socketFd int, recvBuf []byte) ([]byte, error) {
	// Читаем ответ
	n, _, coreFlags, _, err := syscall.Recvmsg(socketFd, recvBuf, nil, 0)
	if err != nil {
		// EAGAIN or EWOULDBLOCK
		// 		The  socket  is marked nonblocking and the receive operation would block, or a receive timeout had been
		// 		set and the timeout expired before data was received.  POSIX.1 allows either error to be  returned  for
		// 		this  case,  and  does  not  require  these constants to have the same value, so a portable application
		// 		should check for both possibilities.
		// игнорируем и обрабатываем выше по стеку

		// EBADF  The argument sockfd is an invalid file descriptor.
		if errors.Is(err, syscall.EBADF) {
			return nil, models.ErrSocketBadFD
		}

		// ECONNREFUSED
		// 		A remote host refused to allow the network connection (typically because it  is  not  running  the  re‐
		// 		quested service).
		if errors.Is(err, syscall.ECONNREFUSED) {
			return nil, models.ErrConnectionRefused
		}

		// EFAULT The receive buffer pointer(s) point outside the process's address space.
		if errors.Is(err, syscall.EFAULT) {
			return nil, models.ErrSpaceAddress
		}

		// EINTR  The receive was interrupted by delivery of a signal before any data was available; see signal(7).
		if errors.Is(err, syscall.EINTR) {
			return nil, models.ErrSignalInterruption
		}

		// EINVAL Invalid argument passed.
		if errors.Is(err, syscall.EINVAL) {
			return nil, models.ErrBadValue
		}

		// ENOMEM Could not allocate memory for recvmsg().
		if errors.Is(err, syscall.ENOMEM) {
			return nil, models.ErrNoMemory
		}

		// ENOTCONN
		// 		The socket is associated with a connection-oriented protocol and has not been connected (see connect(2)
		// 		and accept(2)).
		if errors.Is(err, syscall.ENOTCONN) {
			return nil, models.ErrNotConnected
		}

		// ENOTSOCK The file descriptor sockfd does not refer to a socket
		if errors.Is(err, syscall.ENOTSOCK) {
			return nil, models.ErrSocketBadFD
		}
		
		return nil, fmt.Errorf("receiving err: %w", err)
	}

	// Проверяем флаги ядра
	// Проверка, все ли данные влезли в буфер получения
	if coreFlags & syscall.MSG_TRUNC != 0 {
		return recvBuf, models.ErrRecvMsgTrunc
	}
	// Доп проверка, не должна срабатывать
	if coreFlags & syscall.MSG_CTRUNC != 0 {
		return nil, models.ErrRecvMsgCTrunc
	}
	
	// Соединение закрыто сервером
	if n == 0 {
		return nil, models.ErrConnectionClosed
	}

	return recvBuf[:n], nil
}