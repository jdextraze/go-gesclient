package internal

import (
	"fmt"
	log "github.com/jdextraze/go-gesclient/logger"
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
			panic(fmt.Errorf("No handler registered for message %s", msgID))
		}
		if err := msgHandler(msg); err != nil {
			log.Errorf("Error handling %s: %v", msgID, err) // TODO panic?
		}
	}
}
