package util

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ShutdownHook func()

type ShutdownHandler struct {
	incoming chan os.Signal
	hooks    []ShutdownHook
	mutex    sync.Mutex
	done     chan struct{}
}

func NewShutdownHandler() *ShutdownHandler {
	h := &ShutdownHandler{
		incoming: make(chan os.Signal, 1),
		done:     make(chan struct{}, 1),
	}

	signal.Notify(h.incoming, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT)

	go func() {

		defer func() {
			h.mutex.Lock()
			for _, hook := range h.hooks {
				hook()
			}
			h.mutex.Unlock()
			h.done <- struct{}{}
		}()

		for {
			switch <-h.incoming {

			case syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT:
				return
			}
		}
	}()

	return h
}

func (s *ShutdownHandler) RegisterHook(h ShutdownHook) {
	s.mutex.Lock()
	s.hooks = append(s.hooks, h)
	s.mutex.Unlock()
}

func (s *ShutdownHandler) Wait() {
	<-s.done
}

func (s *ShutdownHandler) TriggerAndWait() {
	defer s.Wait()
	s.incoming <- syscall.SIGTERM
}


