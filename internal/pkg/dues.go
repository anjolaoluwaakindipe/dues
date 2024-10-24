package dues

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/anjolaoluwaakindipe/dues/internal/config"
	"github.com/anjolaoluwaakindipe/dues/internal/debounce"
	"github.com/anjolaoluwaakindipe/dues/internal/filewatcher"
	"github.com/anjolaoluwaakindipe/dues/internal/log"
	"github.com/anjolaoluwaakindipe/dues/internal/pattern"
	"github.com/anjolaoluwaakindipe/dues/internal/process"
	"github.com/anjolaoluwaakindipe/dues/internal/utils"
)

type DuesConfig struct {
	Commands   []string
	ConfigPath string
}

func RunDues(duesConfig DuesConfig) error {
	commands := duesConfig.Commands
	configPath := duesConfig.ConfigPath
	if len(commands) >= 2 {

		return errors.New("Can not process multiple commands as of now!")
	}

	command := commands[0]

	// Read Json file to get all commands available
	var userConfig config.UserConfig
	err := config.ReadConfigFile(configPath, &userConfig)

	if err != nil {
		return errors.New(fmt.Sprintf("An error occured while opening the config file: %v", err))
	}

	userConfig.Process(configPath)

	selectedCommand, err := userConfig.GetCommand(command)

	if err != nil {
		return err
	}

	watcher, err := filewatcher.NewDefaultWatcher()
	if err != nil {
		return fmt.Errorf("could not initialize file watcher: %w", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	d := debounce.NewDebouncer()

	// gorountine to detect file changes, creation, and deletion and perform commands repectively
	go duesProcessLoop(d, selectedCommand, &wg, watcher, sigs)

	wg.Wait()

	return nil
}

func duesProcessLoop(d *debounce.Debouncer, command *process.Command, wg *sync.WaitGroup, watcher filewatcher.Watcher, sigs chan os.Signal) {

	// Callback function to help add detected files to watcher
	addFilesToWatcher := func(path string) {
		watcher.Add(path)
	}

	utils.WalkSubdirectories(command.Cwd, addFilesToWatcher)

	if d == nil || wg == nil || watcher == nil {
		log.Logger.Error("Nil pointer passed to RunDues")
		return
	}

	defer func() {
		d.Cancel()
		watcher.Close()
		wg.Done()
	}()

	var cancellation *context.CancelFunc

	preCtx, done := context.WithTimeout(context.Background(), 15*time.Second)
	err := command.LaunchPreCommand(preCtx)
	if err != nil {
		log.Logger.Error("An error occured launching precommand field:", slog.String("error", err.Error()))
	}
	done()
	ctx, done := context.WithCancel(context.Background())
	go d.Start(1*time.Second, func() {
		if err := command.LaunchCommand(ctx); err != nil {
			log.Logger.Error(fmt.Sprintf("An error occured launching command field: %v", err))
		}
	})

	cancellation = &done
  fileWathcherChannel := watcher.Events()

	for {
		select {
		case event, ok := <-fileWathcherChannel:
			if !ok {
				log.Logger.Error(fmt.Sprintf("An error occured while watching files belonging to command %s", command.Name))
				return
			}
			if event.Has(filewatcher.Write) {
				log.Logger.Debug(fmt.Sprintf("Name of edited event is %v", event.Name()))

				if pattern.Match(event.Name(), command.Ignore) && !pattern.Match(event.Name(), command.Include) {
					continue
				}

				hasReset := d.Reset()
				if !hasReset {
					go d.Start(3*time.Second, func() {
						if cancellation != nil {
							(*cancellation)()
						}
						comCtx, done := context.WithCancel(context.Background())
						cancellation = &done
						err = command.LaunchCommand(comCtx)
						if err != nil {
							log.Logger.Error(fmt.Sprintf("An error occured launching command field: %v", err))
						}
					})
				}
			}
			if event.Has(filewatcher.Create) {
				// We assume that files would already been watched by a
				// specific directory
				if utils.IsDir(event.Name()) {
					utils.WalkSubdirectories(event.Name(), addFilesToWatcher)
				}
			}
			if event.Has(filewatcher.Remove) {
				watcher.Remove(event.Name())
			}
		case err, ok := <-watcher.Errors():
			if !ok {
				return
			}
			log.Logger.Error(fmt.Sprintf("An error occured while wathcing files: %v", err))
		case <-sigs:
			if cancellation != nil {
				(*cancellation)()
			}
			postCtx, done := context.WithTimeout(context.Background(), 15*time.Second)
			err := command.LaunchPostCommand(postCtx)
			if err != nil {
				log.Logger.Error(fmt.Sprintf("An error occured launching post command field: %v", err))
			}
			done()
			return
		}
	}
}
