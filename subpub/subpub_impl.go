package subpub

import (
    "context"
    "sync"
)

type subscription struct {
    unsubscribe func()
}

func (s *subscription) Unsubscribe() {
    s.unsubscribe()
}

type subPub struct {
    mu          sync.RWMutex
    subscribers map[string][]chan interface{}
}

func NewSubPub() SubPub {
    return &subPub{
        subscribers: make(map[string][]chan interface{}),
    }
}

func (sp *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
    ch := sp.addSubscriber(subject)
    sp.startMessageHandler(ch, cb)

    return &subscription{
        unsubscribe: func() {
            sp.removeSubscriber(subject, ch)
        },
    }, nil
}

func (sp *subPub) addSubscriber(subject string) chan interface{} {
    sp.mu.Lock()
    defer sp.mu.Unlock()

    ch := make(chan interface{}, 100)
    sp.subscribers[subject] = append(sp.subscribers[subject], ch)
    return ch
}

func (sp *subPub) startMessageHandler(ch chan interface{}, cb MessageHandler) {
    go func() {
        for msg := range ch {
            cb(msg)
        }
    }()
}

func (sp *subPub) removeSubscriber(subject string, ch chan interface{}) {
    sp.mu.Lock()
    defer sp.mu.Unlock()

    for i, subscriber := range sp.subscribers[subject] {
        if subscriber == ch {
            sp.subscribers[subject] = append(sp.subscribers[subject][:i], sp.subscribers[subject][i+1:]...)
            break
        }
    }
    close(ch)
}

func (sp *subPub) Publish(subject string, msg interface{}) error {
    sp.mu.RLock()
    defer sp.mu.RUnlock()

    for _, ch := range sp.subscribers[subject] {
        sp.sendMessage(ch, msg)
    }
    return nil
}

func (sp *subPub) sendMessage(ch chan interface{}, msg interface{}) {
    select {
    case ch <- msg:
    default:
		// Пропускаем сообщение, если канал переполнен
		// Из условия: Один медленный подписчик не должен тормозить остальных.
    }
}

func (sp *subPub) Close(ctx context.Context) error {
    done := make(chan struct{})
    go sp.closeAllSubscribers(done)

    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-done:
        return nil
    }
}

func (sp *subPub) closeAllSubscribers(done chan struct{}) {
    sp.mu.Lock()
    defer sp.mu.Unlock()

    for _, channels := range sp.subscribers {
        for _, ch := range channels {
            close(ch)
        }
    }
    sp.subscribers = nil
    close(done)
}
