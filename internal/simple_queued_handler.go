package internal

import (
	"reflect"
	"sync/atomic"
	"fmt"
)

type messageHandler func(message message) error

type simpleQueuedHandler struct {
	messageQueue chan message
	handlers map[reflect.Type]messageHandler
	isProcessing int32
}

func newSimpleQueuedHandler(maxQueueSize int) *simpleQueuedHandler {
	return &simpleQueuedHandler{
		messageQueue: make(chan message, maxQueueSize),
		handlers: map[reflect.Type]messageHandler{},
	}
}

func (h *simpleQueuedHandler) RegisterHandler(msg message, handler messageHandler) error {
	h.handlers[reflect.TypeOf(msg)] = handler
	return nil
}

func (h *simpleQueuedHandler) EnqueueMessage(msg message) error {
	h.messageQueue <- msg
	if atomic.CompareAndSwapInt32(&h.isProcessing, 0, 1) {
		go h.processQueue()
	}
	return nil
}

func (h *simpleQueuedHandler) processQueue() {
	for {
		select {
		case msg := <- h.messageQueue:
			msgType := reflect.TypeOf(msg)
			msgHandler, found := h.handlers[msgType]
			if !found {
				panic(fmt.Errorf("No handler registered for message %s", msgType.Name()))
			}
			msgHandler(msg)
		}
		atomic.SwapInt32(&h.isProcessing, 0)
		if len(h.messageQueue) > 0 && atomic.CompareAndSwapInt32(&h.isProcessing, 0, 1) {
			continue
		} else {
			break
		}
	}
}
