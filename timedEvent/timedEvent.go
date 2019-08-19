package timedEvent

import "sync"

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

func (t *TimedEvent) ChanSwitch(finished chan bool, light *sync.Mutex) {
	t.Light.Lock()
	t.Finished <- true
	t.Light.Unlock()
}
