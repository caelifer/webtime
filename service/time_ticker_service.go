package service

import (
	"log"
	"sync"
	"time"
)

type TimeTicker struct {
	m           *sync.RWMutex
	c           *sync.Cond
	currentTime string
}

func NewTimeTicker() *TimeTicker {
	var mtx sync.RWMutex
	ticker := &TimeTicker{
		m: &mtx,
		c: sync.NewCond(mtx.RLocker()),
	}

	// Start updater on the background
	go ticker.start()

	return ticker
}

func (tu *TimeTicker) start() {
	for {
		// Get timestamp each new tick
		tm := <-time.Tick(1 * time.Second)

		// Update current time with exclusive lock
		tu.m.Lock()
		tu.currentTime = tm.Local().Format("15:04:05 MST")
		// log.Printf("TimeUpdater: %s\n", tu.currentTime)
		tu.m.Unlock()

		// Update is complete, signal all clients
		tu.c.Broadcast()
	}
}

func (tu *TimeTicker) Service(quit <-chan bool) <-chan interface{} {
	out := make(chan interface{})

	// Register for periodic updates and run in a separte goroutine
	go func() {
		tu.c.L.Lock()

	LOOP:
		for {
			// Check our state
			select {
			case out <- tu.currentTime:
			case <-quit:
				break LOOP
			default:
			}

			tu.c.Wait()
		}

		// Our client is done. Time to clean-up.
		tu.c.L.Unlock()
		close(out)
		log.Println("TimeTicker: Cleaned up Ticker() resources")
	}()

	return out
}
