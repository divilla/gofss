package fss

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type (
	sessionHandler struct {
		savePath       string
		expireInterval time.Duration
		mu             sync.Mutex
	}
)

func newSessionHandler(cfg SessionStoreConfig) *sessionHandler {
	return &sessionHandler{
		savePath:       cfg.SavePath,
		expireInterval: cfg.ExpireInterval,
	}
}

func (h *sessionHandler) create(id string, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.createSubDirectoryIfNotExists(id); err != nil {
		return err
	}

	filePath := h.getSessionPath(id)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return os.ErrExist
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		slog.Error("create session file failed", slog.String("path", filePath), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (h *sessionHandler) update(id string, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	filePath := h.getSessionPath(id)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	err = f.Truncate(0)
	if err != nil {
		panic(err)
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(data)
	if err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}

	return nil
}

func (h *sessionHandler) delete(id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	filePath := h.getSessionPath(id)
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		slog.Error("delete session file failed", slog.String("path", filePath), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (h *sessionHandler) read(id string) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	filePath := h.getSessionPath(id)
	return os.ReadFile(filePath)
}

func (h *sessionHandler) timestamp(id string) (*time.Time, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	filePath := h.getSessionPath(id)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	atime := fileInfo.Sys().(*syscall.Stat_t).Atim
	unixTime := time.Unix(atime.Sec, atime.Nsec)

	return &unixTime, nil
}

func (h *sessionHandler) purge(prefix string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	dirPath := filepath.Join(h.savePath, prefix+"*")
	subDirs, err := filepath.Glob(dirPath)
	if err != nil && !os.IsNotExist(err) {
		slog.Error("read sessions directory failed", slog.String("path", dirPath), slog.String("error", err.Error()))
		return err
	}

	for _, subDir := range subDirs {
		var filePaths []string
		dirPath = filepath.Join(subDir, "*")
		filePaths, err = filepath.Glob(dirPath)
		if err != nil {
			slog.Error("read sessions subdirectory failed", slog.String("path", dirPath), slog.String("error", err.Error()))
			return err
		}

		for _, filePath := range filePaths {
			var fileInfo os.FileInfo
			fileInfo, err = os.Stat(filePath)
			if err != nil {
				slog.Error("read session file info failed", slog.String("path", filePath), slog.String("error", err.Error()))
				return err
			}

			atime := fileInfo.Sys().(*syscall.Stat_t).Atim
			unixTime := time.Unix(atime.Sec, atime.Nsec)
			if now.Sub(unixTime) > h.expireInterval {
				err = os.Remove(filePath)
				if err != nil && !os.IsNotExist(err) {
					slog.Error("delete session file failed", slog.String("path", filePath), slog.String("error", err.Error()))
					return err
				}
			}
		}
	}

	return nil
}

func (h *sessionHandler) getSubDirectoryName(id string) string {
	return id[0:2]
}

func (h *sessionHandler) getSessionPath(id string) string {
	return filepath.Join(h.savePath, h.getSubDirectoryName(id), id)
}

func (h *sessionHandler) createSubDirectoryIfNotExists(id string) error {
	dirPath := filepath.Join(h.savePath, h.getSubDirectoryName(id))
	_, err := os.Stat(dirPath)
	if errors.Is(err, fs.ErrNotExist) {
		if err = os.Mkdir(dirPath, 0700); err != nil && !os.IsExist(err) {
			slog.Error("session subdirectory create error", slog.String("path", dirPath), slog.String("error", err.Error()))
			return err
		}
	}

	return nil
}
