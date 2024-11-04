package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/anjolaoluwaakindipe/dues/internal/debounce"
	"github.com/anjolaoluwaakindipe/dues/internal/filewatcher"
	"github.com/anjolaoluwaakindipe/dues/internal/log"
	"github.com/anjolaoluwaakindipe/dues/internal/pattern"
	"github.com/anjolaoluwaakindipe/dues/internal/process"
	"github.com/anjolaoluwaakindipe/dues/internal/utils"
)

type DuesCommandRunner struct {
	debouncer *debounce.Debouncer
	command   *process.Command
	watcher   filewatcher.Watcher
}

// addFilesToWatcher includes paths the a filewatcher.Watcher
func (dr *DuesCommandRunner) addFilesToWatcher(path string) {
	dr.watcher.Add(path)
}

// cleanup cleans up the the CommandLoop
func (dr *DuesCommandRunner) cleanUp(wg *sync.WaitGroup) {
	dr.debouncer.Cancel()
	dr.watcher.Close()
	wg.Done()
}

// startMainCommand starts the command field of the process.Command given inside the
// Debouncer
func (dr *DuesCommandRunner) startMainCommand(cancellation *context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	dr.debouncer.Start(100*time.Millisecond, func() {
		if cancellation != nil {
			(*cancellation)()
			cancellation = &cancel
		}
		if err := dr.command.LaunchCommand(ctx); err != nil {
			log.Logger.Error(fmt.Sprintf("An error occured launching command field: %v", err))
		}
	})
}

// CommandLoop is the main loop that watches and manages file events, executes all commands in a
// process.Command, and handles debouncing on file changes
func (dr *DuesCommandRunner) CommandLoop(wg *sync.WaitGroup, sigs chan os.Signal, ctx context.Context) {
	defer dr.cleanUp(wg)
	utils.WalkSubdirectories(dr.command.Cwd, dr.addFilesToWatcher)

	var cancellation *context.CancelFunc

	preCtx, done := context.WithTimeout(context.Background(), 15*time.Second)
	err := dr.command.LaunchPreCommand(preCtx)

	if err != nil {
		log.Logger.Error(fmt.Sprintf("An error occured launching pre-command field: %v", err))
	}
	done()

	go dr.startMainCommand(cancellation)
	eventChannel := dr.watcher.Events()

	for {
		select {
		case event, ok := <-eventChannel:
			if !ok {
				log.Logger.Error(fmt.Sprintf("An error occured while watching files belonging to command %s", dr.command.Name))
			}

			if event.Has(filewatcher.Write) {
				log.Logger.Debug(fmt.Sprintf("Name of edited event is %v", event.Name()))

				if pattern.Match(event.Name(), dr.command.Ignore) && !pattern.Match(event.Name(), dr.command.Include) {
					continue
				}
				eventDelay := 1 * time.Second
				hasReset := dr.debouncer.Reset(&eventDelay)
				if !hasReset {
					go dr.startMainCommand(cancellation)
				}
			}
			if event.Has(filewatcher.Create) {
				// We assume that files would already been watched by a
				// specific directory
				if utils.IsDir(event.Name()) {
					utils.WalkSubdirectories(event.Name(), dr.addFilesToWatcher)
				}
			}
			if event.Has(filewatcher.Remove) {
				dr.watcher.Remove(event.Name())
			}
		case err, ok := <-dr.watcher.Errors():
			if !ok {
				return
			}
			log.Logger.Error(fmt.Sprintf("An error occured while wathcing files: %v", err))
    case <-ctx.Done():
		case <-sigs:
			// Waits for interrupt signal from the os and begins cancellation and clean up process
			if cancellation != nil {
				(*cancellation)()
			}
			postCtx, done := context.WithTimeout(context.Background(), 15*time.Second)
			err := dr.command.LaunchPostCommand(postCtx)
			if err != nil {
				log.Logger.Error(fmt.Sprintf("An error occured launching post command field: %v", err))
			}
			done()
			return
		}
	}
}

func WithDebouncer(d *debounce.Debouncer) DuesRunnerOptions {
	return func(dr *DuesCommandRunner) {
		dr.debouncer = d
	}
}
func WithWatcher(w filewatcher.Watcher) DuesRunnerOptions {
	return func(dr *DuesCommandRunner) {
		dr.watcher = w
	}
}
func WithCommand(c *process.Command) DuesRunnerOptions {
	return func(dr *DuesCommandRunner) {
		dr.command = c
	}
}

type DuesRunnerOptions func(*DuesCommandRunner)

func NewDuesCommandRunner(options ...DuesRunnerOptions) (*DuesCommandRunner, error) {
	runner := DuesCommandRunner{
		debouncer: debounce.NewDebouncer(),
	}

	for _, opt := range options {
		opt(&runner)
	}

	if runner.watcher == nil {
		return nil, errors.New("no filewatcher.Watcher was provided")
	}

	if runner.command == nil {
		return nil, errors.New("no process.Command was provided")
	}

	return &runner, nil
}
