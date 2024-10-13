/*
Copyright © 2024 The Dues Authors
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/anjolaoluwaakindipe/dues/internal/config"
	"github.com/anjolaoluwaakindipe/dues/internal/debounce"
	"github.com/anjolaoluwaakindipe/dues/internal/log"
	"github.com/anjolaoluwaakindipe/dues/internal/pattern"
	"github.com/anjolaoluwaakindipe/dues/internal/process"
	"github.com/anjolaoluwaakindipe/dues/internal/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	configPath    = "dues.json"
	rootCmd       = &cobra.Command{
		Use:   "dues",
		Short: "A live reloading application made to handle multiple tasks concurrently",
		Long: ``,
    Args: cobra.MatchAll(cobra.MinimumNArgs(1)),
		RunE: rootRun,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
    log.Logger.Error(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringVar(&configPath, "config", configPath, "Your dues config path.")
}

// root command execution
func rootRun(cmd *cobra.Command, args []string) error {
  commands := args
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
  

	watcher, err := fsnotify.NewWatcher()


	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	d := debounce.NewDebouncer()

	// gorountine to detect file changes, creation, and deletion and perform commands repectively
	go RunDues(d, selectedCommand, &wg, watcher, sigs)

	wg.Wait()

  return nil
}


func RunDues(d *debounce.Debouncer, command *process.Command, wg *sync.WaitGroup, watcher *fsnotify.Watcher, sigs chan os.Signal) {
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
		log.Logger.Error(fmt.Sprintf("An error occured launching precommand field: %v", err))
	}
	done()
	ctx, done := context.WithCancel(context.Background())
	go d.Start(1*time.Millisecond, func() {
		if err := command.LaunchCommand(ctx); err != nil {
			log.Logger.Error(fmt.Sprintf("An error occured launching command field: %v", err))
		}
	})

	cancellation = &done

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				log.Logger.Debug(fmt.Sprintf("Name of edited event is %v", event.Name))

				if pattern.Match(event.Name, command.Ignore) && !pattern.Match(event.Name, command.Include) {
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
			if event.Has(fsnotify.Create) {
				log.Logger.Debug(fmt.Sprintf("Name of created event is %v", event.Name))
				if utils.IsDir(event.Name) {
					utils.WalkSubdirectories(event.Name, addFilesToWatcher)
				}
			}
			if event.Has(fsnotify.Remove) {
				watcher.Remove(event.Name)
			}
		case err, ok := <-watcher.Errors:
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
