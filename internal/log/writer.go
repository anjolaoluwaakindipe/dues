/*
Copyright © 2024 The Dues Authors
*/
package log

import (
	"fmt"
	"io"
	"time"
)

type line struct {
	Message   string
	Timestamp string
	Severity  string
}

type DuesWriter struct {
	writer        io.Writer
	severity string
	time     string
	command  string
}

func (dw *DuesWriter) Write(p []byte) (int, error) {
	l := &line{
		Message:   string(p),
		Timestamp: dw.time,
		Severity:  dw.severity,
	}
  if dw.time == ""{
   l.Timestamp = time.Now().Format(time.RFC850) 
  }

	data := fmt.Sprintf("%s %s %s", l.Timestamp, l.Severity, l.Message)

	if dw.command != "" {
		data = dw.command + " " + data
	}

	n, err := dw.writer.Write([]byte(data))

	if err != nil {
		return n, err
	}

	if n != len(p) {
		return n, io.ErrShortWrite
	}

	return len(p), nil
}

func NewDuesWriter(writer io.Writer, severity string, command string) *DuesWriter {
 return &DuesWriter{
    writer: writer,
    severity: severity,
    command: command,
  } 
}
