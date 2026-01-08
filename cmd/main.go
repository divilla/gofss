package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/divilla/gofss"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLogger()) // use the default RequestLogger middleware with slog logger
	e.Use(middleware.Recover())       // recover panics as errors for proper error handling

	// Routes
	e.GET("/", hello)

	sh, err := gofss.NewSessionHandler("/tmp/sessions")
	if err != nil {
		panic(err)
	}

	var id string
	data := []byte("hello world")
	for {
		id = gofss.NewHash(8)
		if err = sh.Open(id, data); err == nil {
			break
		} else {
			log.Error(err)
		}
	}

	data = []byte("hello big world")
	err = sh.Write(id, data)
	if err != nil {
		panic(err)
	}

	err = sh.Destroy(id)
	if err != nil {
		panic(err)
	}

	// Start server
	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, fmt.Sprintf("%s\n", gofss.NewHash(8)))
}
