package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type logWriter struct {
	writers []io.Writer
	file    *os.File
}

func newLogWriter() *logWriter {
	lw := &logWriter{}

	consoleVal := os.Getenv("LOG_CONSOLE")
	if !(strings.EqualFold(consoleVal, "false") || consoleVal == "0") {
		lw.writers = append(lw.writers, os.Stdout)
	}

	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", logFile, err)
		} else {
			lw.file = f
			lw.writers = append(lw.writers, f)
		}
	}

	return lw
}

func (lw *logWriter) Write(p []byte) (int, error) {
	for _, w := range lw.writers {
		w.Write(p)
	}
	return len(p), nil
}

func (lw *logWriter) Close() error {
	if lw.file != nil {
		return lw.file.Close()
	}
	return nil
}

func requestLogger(w io.Writer) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		chainErr := c.Next()

		status := c.Response().StatusCode()
		if status == 0 {
			status = fiber.StatusOK
		}

		var level, color string
		switch {
		case status >= 500:
			level = "ERROR"
			color = "\033[31m"
		case status >= 400:
			level = "WARN"
			color = "\033[33m"
		default:
			level = "INFO"
			color = "\033[36m"
		}

		dur := time.Since(start)
		ip := c.IP()
		method := c.Method()
		path := c.Path()

		fmt.Fprintf(
			w,
			"%s %s[%s]%s %3d | %13v | %-7s %s | %s\n",
			start.Format("2006/01/02 15:04:05"),
			color, level, "\033[0m",
			status,
			dur,
			method, path,
			ip,
		)

		return chainErr
	}
}
