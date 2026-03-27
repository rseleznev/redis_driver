package connection

import (
	"errors"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/internal/connection/socket"
)

// New создает новое подключение
func New(opts models.Options) (int, error) {
	// Создаем сокет
	socketFd, err := socket.New()
	if err != nil {
		return 0, err
	}

	// Ставим на отслеживание первичное событие
	err = epoll.InitEventForSocket(socketFd)
	if err != nil {
		return 0, err
	}

	attempt := 1 // счетчик ретраев

	for {
		// Подключаемся
		err = socket.Connect(opts.RedisIp, opts.RedisPort, socketFd)
		if err != nil {
			// Проверяем ошибки, при которых нет смысла делать ретраи
			switch err {
			case models.ErrSocketNoAccess, models.ErrSocketLocalPortInUse, models.ErrSocketNoLocalPorts, models.ErrAddrBadParams,
			models.ErrConnectionInProcess, models.ErrSocketBadFD, models.ErrSignalInterruption:
				return 0, err

			}

			// Делаем ретраи
			if attempt < opts.RetryAmount {
				attempt++
				continue
			} else {
				return 0, models.ErrConnectionRetriesFailed
			}
		}

		// Ждем результат
		epoll.Wait()

		// Проверка события
		err = epoll.ProcessEvent(socketFd)
		if err != nil {
			// Проверяем ошибки, при которых нет смысла делать ретраи
			if errors.Is(err, syscall.ECONNREFUSED) {
				return 0, models.ErrConnectionRefused
			}
			if errors.Is(err, syscall.EHOSTUNREACH) {
				return 0, models.ErrServerUnreachable
			}
			if errors.Is(err, syscall.ENETUNREACH) {
				return 0, models.ErrNetUnreachable
			}
			if errors.Is(err, syscall.EACCES) {
				return 0, models.ErrSocketNoAccess
			}

			// ErrSocketEvent, ErrSocketHUPEvent и ErrSocketRDHUPEvent точно не надо проверять?
			
			// Делаем ретраи
			if attempt < opts.RetryAmount {
				attempt++
				continue
			} else {
				return 0, models.ErrConnectionRetriesFailed
			}
		}
		break
	}

	// Ставим на отслеживание входящие события
	err = epoll.AddIncomeEventForSocket(socketFd)
	if err != nil {
		return 0, err
	}

	return socketFd, nil
}

// Reconnect переподключается к серверу
func Reconnect(opts models.Options, oldSocketFd int) (int, error) {
	// Закрываем старый сокет
	syscall.Close(oldSocketFd)
	
	// // Удаляем события закрытого сокета из списка отслеживания
	// err := epoll.DeleteEventsForSocket(oldSocketFd)
	// if err != nil {
	// 	return 0, err
	// }

	// Создаем и подключаем новый сокет
	newSocketFd, err := New(opts)
	if err != nil {
		return 0, err
	}

	return newSocketFd, nil
}