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

// Start invokess the debouncer callback after a specific delay.
// If the reset method is called before the callback is invoked then
// the delay will be reset. If the Cancel Method is called then the delay will 
// be stopped and the callback will not be invoked
func (self *Debouncer) Start(delay time.Duration, callback func()) {
	self.timer = time.NewTimer(delay)
  defer func(){
    self.started = false
  }()
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
      return 
		}
	}
}

func (self *Debouncer) Reset(delay *time.Duration) bool {
	if !self.started {
		return false
	}
  if delay != nil {
    self.timer = time.NewTimer(*delay)
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
