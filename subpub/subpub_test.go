package subpub

import (
    "context"
    "sync"
    "testing"
)

func TestSubPub(t *testing.T) {
    sp := NewSubPub()
    defer sp.Close(context.Background())

    var wg sync.WaitGroup
    wg.Add(1)

    sub, err := sp.Subscribe("test", func(msg interface{}) {
        if msg.(string) != "hello" {
            t.Errorf("expected 'hello', got '%v'", msg)
        }
        wg.Done()
    })
    if err != nil {
        t.Fatalf("failed to subscribe: %v", err)
    }

    err = sp.Publish("test", "hello")
    if err != nil {
        t.Fatalf("failed to publish: %v", err)
    }

    wg.Wait()
    sub.Unsubscribe()
}
