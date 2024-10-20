/*
Copyright © 2024 The Dues Authors
*/
package log

import (
	"fmt"
	"math/rand"
	"strconv"
)

type StringColor = int

const (
	reset                    = "\033[0m"
	Black        StringColor = 30
	Red          StringColor = 31
	Green        StringColor = 32
	Yellow       StringColor = 33
	Blue         StringColor = 34
	Magenta      StringColor = 35
	Cyan         StringColor = 36
	LightGray    StringColor = 37
	DarkGray     StringColor = 90
	LightRed     StringColor = 91
	LightGreen   StringColor = 92
	LightYellow  StringColor = 93
	LightBlue    StringColor = 94
	LightMagenta StringColor = 95
	LightCyan    StringColor = 96
	White        StringColor = 97
)

func GetRandomStringColor() StringColor{
  colorOptions := []StringColor{ Blue, Magenta, LightBlue, LightMagenta, Red}
  selectedColor := colorOptions[rand.Intn(len(colorOptions))]
  return selectedColor
}

func ColorizeRandom(v string) string {
  selectedColor := GetRandomStringColor()
  return Colorize(selectedColor, v)
}

func Colorize(colorCode StringColor, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}
