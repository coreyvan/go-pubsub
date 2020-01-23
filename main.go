package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	var numClients = 10
	var wg sync.WaitGroup

	ps := NewPubsub()

	for i := 0; i < numClients; i++ {
		ch := ps.Subscribe("sharks")
		wg.Add(1)
		go Listen(i, ch, &wg)
	}

	for i := 0 + numClients; i < numClients*2; i++ {
		ch := ps.Subscribe("fish")
		wg.Add(1)
		go Listen(i, ch, &wg)
	}

	ps.Publish("sharks", "holy cow sharks")
	time.Sleep(3 * time.Second)
	ps.Publish("sharks", "look out sharks")
	ps.Publish("fish", "oh cool fish")
	time.Sleep(3 * time.Second)
	ps.Publish("sharks", "ok no more sharks")
	ps.Close()
	fmt.Println("Pubsub closed signal sent, waiting for clients")
	wg.Wait()
	fmt.Println("all clients done, ending")
}

// Pubsub struct defining a pubsub
type Pubsub struct {
	mu     sync.RWMutex
	subs   map[string][]chan string
	closed bool
}

// NewPubsub returns a new Pubsub
func NewPubsub() *Pubsub {
	ps := &Pubsub{}
	ps.subs = make(map[string][]chan string)
	return ps
}

// Subscribe subscribes a client to a topic
func (ps *Pubsub) Subscribe(topic string) <-chan string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan string, 1)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

// Publish publishes messages to a topic
func (ps *Pubsub) Publish(topic string, msg string) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return
	}

	for _, ch := range ps.subs[topic] {
		ch <- msg
	}
}

// Close closes a pubsub
func (ps *Pubsub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, subs := range ps.subs {
			for _, ch := range subs {
				close(ch)
			}
		}
	}
}

// Listen listens on channel and prints
func Listen(w int, ch <-chan string, wg *sync.WaitGroup) {
	for msg := range ch {
		source := rand.NewSource(time.Now().UnixNano())
		random := rand.New(source)
		delay := random.Intn(2000)
		time.Sleep(time.Duration(delay) * time.Millisecond)
		fmt.Printf("[%d] Delay: %d - Got message: %s\n", w, delay, msg)
	}
	fmt.Printf("[%d] Channel closed\n", w)
	wg.Done()
}
