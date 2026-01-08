package gofss

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

type (
	SessionHandler struct {
		savePath string
		mu       sync.Mutex
	}
)

func NewSessionHandler(savePath string) (*SessionHandler, error) {
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		err = os.Mkdir(savePath, 0755)
		if err != nil {
			slog.Error("make session sub dir error: %w", err)
			return nil, err
		}
	}

	return &SessionHandler{savePath: savePath}, nil
}

func (h *SessionHandler) Read(id string) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var file *os.File
	var err error

	dirPath := filepath.Join(h.savePath, id[0:2])
	if _, err := os.Stat(dirPath); err != nil {
		slog.Error("open session dir: '%s' error: %w", dirPath, err)
		return nil, err
	}

	filePath := filepath.Join(dirPath, id)
	if file, err = os.Open(filepath.Join(filePath)); err != nil {
		slog.Error("open session file: '%s error: %w", filePath, err)
		return nil, err
	}
	defer func() {
		if err = file.Close(); err != nil {
			slog.Error("close session file error: %w", err)
		}
	}()

	b, err := io.ReadAll(file)
	if err != nil {
		slog.Error("read session file error: %w", err)
		return nil, err
	}

	return b, nil
}

func (h *SessionHandler) Open(id string, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	dirPath := filepath.Join(h.savePath, id[0:2])
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.Mkdir(dirPath, 0755)
		if err != nil {
			slog.Error("make session dir: '%s' error: %w", dirPath, err)
			return err
		}
	}
	if _, err := os.Stat(dirPath); err != nil {
		slog.Error("open session dir: '%s' error: %w", dirPath, err)
		return err
	}

	filePath := filepath.Join(dirPath, id)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		err = errors.New("file already exists")
		slog.Error("session file: '%s' error: %w", filePath, err)
		return err
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		slog.Error("write session file: '%s' error: %w", filePath, err)
		return err
	}

	return nil
}

func (h *SessionHandler) Write(id string, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	dirPath := filepath.Join(h.savePath, id[0:2])
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.Mkdir(dirPath, 0755)
		if err != nil {
			slog.Error("make session dir: '%s' error: %w", dirPath, err)
			return err
		}
	}

	filePath := filepath.Join(dirPath, id)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		slog.Error("write session file: '%s' error: %w", filePath, err)
		return err
	}

	return nil
}

func (h *SessionHandler) Destroy(id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	dirPath := filepath.Join(h.savePath, id[0:2])
	if _, err := os.Stat(dirPath); err != nil {
		slog.Error("open session dir '%s' error: %w", dirPath, err)
		return err
	}

	filePath := filepath.Join(dirPath, id)
	if err := os.Remove(filePath); err != nil {
		slog.Error("remove session file: '%s' error: %w", filePath, err)
		return err
	}

	return nil
}
