package config

import (
	"errors"
	"fmt"

	"github.com/anjolaoluwaakindipe/dues/internal/process"
)

type UserConfig struct {
	Commands map[string]*process.Command
}

// Takes the configuration given and uses it to help validate and process
// the internal command
func (uc *UserConfig) Process(configPath string) error {
	for _, v := range uc.Commands {
		if err := v.Process(configPath); err != nil {
			return err
		}
	}
	return nil
}

// Checks whether command exists in configuration
func (uc *UserConfig) DoesCommandExist(command string) bool {
  _, exists := uc.Commands[command]
  return exists
} 

// Gets a specific command. If command does not exists in configuration then
// An error is returned.
func (uc *UserConfig) GetCommand(command string) (*process.Command, error){
  commandStruct, exists := uc.Commands[command]

  if !exists {
    return nil, errors.New(fmt.Sprintf("Command '%v' does not exists in this config", command))
  }

  commandStruct.Name = command

  return commandStruct, nil
}
