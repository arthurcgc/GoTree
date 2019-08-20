package timedEvent

import (
	"fmt"
	"sync"
	"time"
)

type TimedEvent struct {
	Light    *sync.Mutex
	Finished chan bool
	Wg       *sync.WaitGroup
}

func (t *TimedEvent) IsEmpty() bool {
	return t == nil
}

func NewTimedEvent() *TimedEvent {
	finished := make(chan bool, 1)
	light := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	tEvent := &TimedEvent{
		Light:    light,
		Finished: finished,
		Wg:       wg}

	return tEvent

}

func (t *TimedEvent) CheckReceiveSignalNoHang() bool {
	select {
	case <-t.Finished:
		return true
	default:
		return false
	}
}

func (t *TimedEvent) Sleeping(dur int) {
	seconds := time.Duration(dur) * time.Second
	select {
	case <-t.Finished:
		{
			t.Wg.Done()
			return
		}
	case <-time.After(seconds):
		{
			fmt.Printf("\nProgram timed out!\n")
			t.ChanSwitch()
			t.Wg.Done()
			return
		}
	}
}

func (t *TimedEvent) ChanSwitch() {
	t.Light.Lock()
	t.Finished <- true
	t.Light.Unlock()
}
