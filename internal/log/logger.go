package log

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func InitLogger() {
	if Logger == nil {
		level := slog.LevelVar{}
		level.Set(slog.LevelDebug)
		opts := slog.HandlerOptions{
			Level: &level,
		}
		handler := slog.NewTextHandler(os.Stdout, &opts)
		Logger = slog.New(handler)
	}
}

func init(){
  InitLogger()  
}

