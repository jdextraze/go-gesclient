package gesclient

import (
    "github.com/satori/go.uuid"
    . "bitbucket.org/jdextraze/go-gesclient/protobuf"
)

type SubscriptionResult struct {
    conn          *connection
    correlationId uuid.UUID
    Confirmation  *SubscriptionConfirmation
    Events        chan *StreamEventAppeared
    unsubscribe   chan *SubscriptionDropped
}

func (s *SubscriptionResult) Unsubscribe() (*SubscriptionDropped, error) {
    ch, err := s.UnsubscribeAsync()
    if err != nil {
        return nil, err
    }
    return <-ch, err
}

func (s *SubscriptionResult) UnsubscribeAsync() (chan *SubscriptionDropped, error) {
    if err := s.conn.assertConnected(); err != nil {
        return nil, err
    }
    s.unsubscribe = make(chan *SubscriptionDropped)
    return s.unsubscribe, s.conn.sendCommand(
        tcpCommand_UnsubscribeFromStream,
        &UnsubscribeFromStream{},
        nil,
        s.correlationId,
    )
}
