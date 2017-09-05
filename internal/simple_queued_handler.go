package internal

import (
	"fmt"
)

type messageHandler func(message message) error

type simpleQueuedHandler struct {
	messageQueue chan message
	handlers     map[int]messageHandler
}

func newSimpleQueuedHandler() *simpleQueuedHandler {
	h := &simpleQueuedHandler{
		messageQueue: make(chan message, 65536), // TODO buffer size
		handlers:     map[int]messageHandler{},
	}
	go h.processQueue()
	return h
}

func (h *simpleQueuedHandler) RegisterHandler(msg message, handler messageHandler) error {
	h.handlers[msg.MessageID()] = handler
	return nil
}

func (h *simpleQueuedHandler) EnqueueMessage(msg message) error {
	h.messageQueue <- msg
	return nil
}

func (h *simpleQueuedHandler) processQueue() {
	for msg := range h.messageQueue {
		msgID := msg.MessageID()
		msgHandler, found := h.handlers[msgID]
		if !found {
			panic(fmt.Sprintf("No handler registered for message %d", msgID))
		}
		if err := msgHandler(msg); err != nil {
			log.Errorf("Error handling %d: %v", msgID, err) // TODO panic?
		}
	}
}
