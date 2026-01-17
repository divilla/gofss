package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/divilla/gofss"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLogger()) // use the default RequestLogger middleware with slog logger
	e.Use(middleware.Recover())       // recover panics as errors for proper error handling

	// Routes
	e.GET("/", hello)

	config := fss.NewSessionStoreConfig()
	ss, err := fss.NewSessionStore(config)
	if err != nil {
		panic(err)
	}

	data := []byte("Hello World\n")
	id := ss.Create(data)
	fmt.Println(id)

	id = "hOJDR4Hutl98X0oxxKg3XPMw9bxH6VZHPZ8KiU-92cEmTdZYgl8t_CjC35_an1C2FuPbuWazy2RdC3T0"
	var unixTime *time.Time
	data, err = ss.Read(id)
	fmt.Println("read1", string(data), unixTime, err)

	<-time.After(10 * time.Second)

	data, err = ss.Read(id)
	fmt.Println("read2", string(data), unixTime, err)

	os.Exit(0)

	data = []byte("Hello New World\n")
	err = ss.Update(id, data)
	if err != nil {
		panic(err)
	}

	ss.Delete(id)

	// Start server
	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, fmt.Sprintf("%s\n", "whatever"))
}
