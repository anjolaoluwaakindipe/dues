package dues

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/anjolaoluwaakindipe/dues/internal/config"
	"github.com/anjolaoluwaakindipe/dues/internal/debounce"
	"github.com/anjolaoluwaakindipe/dues/internal/filewatcher"
	"github.com/anjolaoluwaakindipe/dues/internal/process"
	"github.com/anjolaoluwaakindipe/dues/internal/runner"
)

type DuesConfig struct {
	Commands   []string
	ConfigPath string
}

func RunDues(duesConfig DuesConfig) error {
	commands := duesConfig.Commands
	configPath := duesConfig.ConfigPath

	// Read Json file to get all commands available
	var userConfig config.UserConfig
	err := config.ReadConfigFile(configPath, &userConfig)

	if err != nil {
		return errors.New(fmt.Sprintf("An error occured while opening the config file: %v", err))
	}

	userConfig.Process(configPath)

	var commandList []*process.Command

	for _, currCommand := range commands {
		selectedCommand, err := userConfig.GetCommand(currCommand)
		if err != nil {
			return err
		}

		commandList = append(commandList, selectedCommand)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
  backgroundCtx, cancel := context.WithCancel(context.Background())
  var commandListErr error = nil 
	for _, currCommand := range commandList {

		watcher, err := filewatcher.NewDefaultWatcher()
		if err != nil {
      cancel()
			commandListErr = fmt.Errorf("could not initialize file watcher: %w", err)
      break
		}

		d := debounce.NewDebouncer()

		// gorountine to detect file changes, creation, and deletion and perform commands repectively
		commandRunner, err := runner.NewDuesCommandRunner(
			runner.WithCommand(currCommand),
			runner.WithWatcher(watcher),
			runner.WithDebouncer(d),
		)

		if err != nil {
      cancel()
			commandListErr = fmt.Errorf("an error occurred initializing runner: %w", err)
      break
		}

		wg.Add(1)
		go commandRunner.CommandLoop(&wg, sigs, backgroundCtx)

	}

	wg.Wait()
  cancel()

  if commandListErr != nil {
    return commandListErr
  }

	return nil
}
