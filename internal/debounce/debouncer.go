package debounce

import "time"

type DebounceSignal int

const (
	reset DebounceSignal = iota
	cancel
)

type Debouncer struct {
	delay      time.Duration
	signalChan chan DebounceSignal
	started    bool
	timer      *time.Timer
}

func NewDebouncer() *Debouncer {
	return &Debouncer{
		signalChan: make(chan DebounceSignal),
		started:    false,
	}
}

func (self *Debouncer) Start(delay time.Duration, callback func()) {
	self.timer = time.NewTimer(delay)
	for {
		self.started = true
		select {
		case sig := <-self.signalChan:
			if sig == reset {
				self.timer.Reset(delay)
			} else if sig == cancel {
				self.timer.Stop()
				return
			}
		case <-self.timer.C:
			callback()
		}
	}
}

func (self *Debouncer) Reset() bool {
	if !self.started {
		return false
	}
	self.signalChan <- reset
	return true
}

func (self *Debouncer) Cancel() bool {
	if !self.started {
		return false
	}
	self.signalChan <- cancel
	return true
}
