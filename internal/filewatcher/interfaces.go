package filewatcher

type Event interface {
	Has(Operation) bool // if the event has the specific operation
	Name() string
	Operation() Operation
}

type Watcher interface {
	Add(path string) error
	Close() error
	Remove(path string) error
	Events() chan Event
  Errors() chan error
}

type Operation uint32

func (o Operation) Has(op Operation) bool {
  return o&op != 0
}

const (

	// A new pathname was created.
	Create Operation = 1 << iota

	// The pathname was written to; this does *not* mean the write has finished,
	// and a write can be followed by more writes.
	Write

	// The path was removed; any watches on it will be removed. Some "remove"
	// operations may trigger a Rename if the file is actually moved (for
	// example "remove to trash" is often a rename).
	Remove

	// The path was renamed to something else; any watched on it will be
	// removed.
	Rename

	// File attributes were changed.
	//
	// It's generally not recommended to take action on this event, as it may
	// get triggered very frequently by some software. For example, Spotlight
	// indexing on macOS, anti-virus software, backup software, etc.
	Chmod
)
