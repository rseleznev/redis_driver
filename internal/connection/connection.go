package connection

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/internal/connection/socket"
)

var (
	ErrConnectionRetriesFailed = errors.New("redis_driver: all connection retries failed")
	ErrConnectionRefused = errors.New("redis_driver: connection is refused")
	ErrServerUnreachable = errors.New("redis_driver: server is unreachable")
	ErrNetUnreachable = errors.New("redis_driver: network is unreachable")
)

// New создает новое подключение
func New(opts models.Options) (int, error) {
	// Создаем сокет
	socketFd, err := socket.New()
	if err != nil {
		return 0, err
	}
	attempt := 1

	for {
		// Подключаемся
		err = socket.Connect(opts.RedisIp, opts.RedisPort, socketFd)
		if err != nil {
			// Делаем ретраи
			if attempt < opts.RetryAmount {
				attempt++
				continue
			} else {
				return 0, ErrConnectionRetriesFailed
			}
		}

		if attempt == 1 {
			// Ставим на отслеживание первичное событие
			err = epoll.InitEventForSocket(socketFd)
			if err != nil {
				fmt.Println(err)
			}	
		}
		
		// Ждем результат
		epoll.Wait()

		// Проверка события
		err = epoll.ProcessEvent(socketFd)
		if err != nil {
			// Проверяем ошибки, при которых нет смысла делать ретраи
			if errors.Is(err, syscall.ECONNREFUSED) {
				return 0, ErrConnectionRefused
			}
			if errors.Is(err, syscall.EHOSTUNREACH) {
				return 0, ErrServerUnreachable
			}
			if errors.Is(err, syscall.ENETUNREACH) {
				return 0, ErrNetUnreachable
			}
			if errors.Is(err, syscall.EACCES) {
				return 0, socket.ErrSocketNoAccess
			}
			
			// Делаем ретраи
			if attempt < opts.RetryAmount {
				attempt++
				continue
			} else {
				return 0, ErrConnectionRetriesFailed
			}
		}
		break
	}
	fmt.Println("Сокет подключен!")

	// Ставим на отслеживание входящие события
	err = epoll.AddIncomeEventForSocket(socketFd)
	if err != nil {
		fmt.Println(err)
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