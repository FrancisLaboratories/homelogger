package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

func requestLogger() fiber.Handler {
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

		fmt.Printf(
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
