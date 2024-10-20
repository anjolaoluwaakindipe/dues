/*
Copyright © 2024 The Dues Authors
*/
package log

import (
	"log/slog"
	"os"
)

// Default Logger for the application
var Logger *slog.Logger

// initializes the default logger
func initDefaultLogger() {
	if Logger == nil {
		level := slog.LevelVar{}
		level.Set(slog.LevelDebug)
		opts := slog.HandlerOptions{
			Level: &level,
		}
		handler := NewDuesHandler(os.Stdout, &opts)
		Logger = slog.New(handler)
	}
}

func init(){
  initDefaultLogger()  
}

