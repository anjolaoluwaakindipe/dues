package process

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anjolaoluwaakindipe/dues/internal/log"
)

type Command struct {
	Command     string
	PreCommand  string
	PostCommand string
	Cwd         string
	Name        string
	Ignore      []string
	Include     []string
	Color       log.StringColor
}

// Validates the command structure
func (c *Command) Process(configPath string) error {
  c.Color = log.GetRandomStringColor()

	if err := c.processCommand(); err != nil {
		return err
	}

	if err := c.processPreCommand(); err != nil {
		return err
	}

	if err := c.processPostCommand(); err != nil {
		return err
	}

	if err := c.processCwd(configPath); err != nil {
		return err
	}

	return nil
}

// Validates the command field
func (c *Command) processCommand() error {
	c.Command = strings.TrimSpace(c.Command)
	if len(c.Command) == 0 {
		return errors.New(fmt.Sprintf("Command '%v' has an empty command field", c.Name))
	}
	return nil
}

// Validates whether the cwd exists. If the Cwd field is a valid absolute path
// then no error is returned. If not it is assumed that the Cwd path is a relative path
// with respective to the configPath. Thus the Cwd field with be checked to validate that
// it is a valid relative path
func (c *Command) processCwd(configPath string) error {
	c.Cwd = strings.TrimSpace(c.Cwd)

	if !filepath.IsAbs(c.Cwd) {
		return nil
	}

	configDirPath := filepath.Dir(configPath)
	cwd, err := filepath.Abs((filepath.Join(configDirPath, c.Cwd)))

	c.Cwd = cwd

	if err != nil {
		return errors.New(fmt.Sprintf("Could not find path specified in command '%v' cwd's field.", c.Name))
	}

	return nil
}

// Validates pre command
func (c *Command) processPreCommand() error {
	c.PreCommand = strings.TrimSpace(c.PreCommand)
	return nil
}

// Validates post command
func (c *Command) processPostCommand() error {
	c.PostCommand = strings.TrimSpace(c.PostCommand)
	return nil
}

// Converts comand string to slice, delimited by whitespaces
func (c *Command) commandSlice() []string {
	commandAsSlice := strings.Fields(c.Command)
	return commandAsSlice
}

// Converts pre comand string to slice, delimited by whitespaces
func (c *Command) preCommandSlice() []string {
	commandAsSlice := strings.Fields(c.PreCommand)
	return commandAsSlice
}

// Converts post command string to slice, delimited by whitespaces
func (c *Command) postCommandSlice() []string {
	commandAsSlice := strings.Fields(c.PostCommand)
	return commandAsSlice
}

// Launches pre command
func (c *Command) LaunchPreCommand(ctx context.Context) error {
	pcs := c.preCommandSlice()
	return c.runCmd(pcs, ctx)
}

// Launches command
func (c *Command) LaunchCommand(ctx context.Context) error {
	cs := c.commandSlice()
	return c.runCmd(cs, ctx)
}

// Launches post command
func (c *Command) LaunchPostCommand(ctx context.Context) error {
	pcs := c.postCommandSlice()
	return c.runCmd(pcs, ctx)
}

// Runs a specific command. This is a blocking operation until said command finishes execution,
// thus run in separate goroutine for if concurrency is needed
func (c *Command) runCmd(command []string, ctx context.Context) error {
	if len(command) == 0 {
		return errors.New("length of command string slice is zero")
	}
	var cmd *exec.Cmd

	if len(command) == 1 {
		cmd = exec.CommandContext(ctx, command[0])
	} else {
		cmd = exec.CommandContext(ctx, command[0], command[1:]...)
	}
	cmd.Dir = c.Cwd
	cmd.Stderr = log.NewDuesWriter(os.Stderr, log.Colorize(log.LightRed, slog.LevelInfo.String()), log.Colorize(c.Color, c.Name))
	cmd.Stdout = log.NewDuesWriter(os.Stdout, log.Colorize(log.LightCyan, slog.LevelInfo.String()), log.Colorize(c.Color, c.Name))

	err := cmd.Start()

	if err != nil {
		return err
	}

	cmd.Wait()
	return nil
}
