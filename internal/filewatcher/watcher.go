package filewatcher

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
)

// DefaultWatcher this a decorator struct that utilizes
// the fsNotify watcher under the hood
type DefaultWatcher struct {
  watcher *fsnotify.Watcher
}

func (dw *DefaultWatcher) Add(path string ) error {
  return dw.watcher.Add(path)
}

func (dw *DefaultWatcher) Close() error {
  return dw.watcher.Close()
}

func (dw *DefaultWatcher) Remove(path string) error {
  return dw.watcher.Remove(path)
}

func (dw *DefaultWatcher) Errors() chan error {
  return dw.watcher.Errors
}

func (dw *DefaultWatcher) Events() chan Event {
  fsnotifyEventChan := dw.watcher.Events
  localEventChannel := make(chan Event)


  go func(){
    for {
      fsnotifyEvent, ok  := <- fsnotifyEventChan
      if !ok{
        close(localEventChannel)
        return 
      }

      event := DefaultFileEvent{}
      event.EventName = fsnotifyEvent.Name
      if fsnotifyEvent.Has(fsnotify.Create){
        event.Op |= Create
      }
      if fsnotifyEvent.Has(fsnotify.Write){
        event.Op |= Write
      }
      if fsnotifyEvent.Has(fsnotify.Remove){
        event.Op |= Remove
      }
      if fsnotifyEvent.Has(fsnotify.Rename){
        event.Op |= Remove
      }
      if fsnotifyEvent.Has(fsnotify.Chmod){
        event.Op |= Chmod
      }

      localEventChannel <- &event
    }
  }()

  return localEventChannel
}


type DefaultFileEvent struct {
  Op Operation
  EventName string
}

func (dfe *DefaultFileEvent) Operation() Operation {
  return dfe.Op
}

func (dfe *DefaultFileEvent) Has(op Operation) bool {
  return dfe.Op.Has(op)
}

// File path the operation happend on 
func (dfe *DefaultFileEvent) Name() string {
  return dfe.EventName
}



func NewDefaultWatcher() (*DefaultWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
  if err != nil {
    return nil, fmt.Errorf("an error occured while spawning default watcher: %w", err)
  }
  return &DefaultWatcher{
    watcher: watcher,
  }, nil 
}

