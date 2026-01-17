package fss

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

var (
	URL64   = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	rndSize = int64(64 * 64 * 64 * 64 * 64 * 64 * 64 * 64 * 64 * 64)
)

type (
	SessionStore struct {
		handlerMap    map[string]*sessionHandler
		hashSize      int
		purgeInterval time.Duration
	}

	SessionStoreConfig struct {
		SavePath       string
		HashSize       int
		ExpireInterval time.Duration
		PurgeInterval  time.Duration
	}
)

func NewSessionStoreConfig() *SessionStoreConfig {
	workDir, _ := os.Getwd()
	return &SessionStoreConfig{
		SavePath:       filepath.Join(workDir, "sessions"),
		HashSize:       8,
		ExpireInterval: time.Hour * 24 * 365,
		PurgeInterval:  time.Hour * 24,
	}
}

func NewSessionStore(config *SessionStoreConfig) (*SessionStore, error) {
	fileInfo, err := os.Stat(config.SavePath)
	if errors.Is(err, fs.ErrNotExist) {
		if err = os.Mkdir(config.SavePath, 0700); err != nil {
			return nil, fmt.Errorf("mkdir: '%s' error: %w", config.SavePath, err)
		}
	} else if err != nil {
		slog.Error("creating session directory '%s' error: %w", config.SavePath, err)
		return nil, err
	} else if !fileInfo.IsDir() {
		return nil, errors.New(fmt.Sprintf("ids savePath is not a directory: %s", config.SavePath))
	}

	app := &SessionStore{
		handlerMap:    make(map[string]*sessionHandler),
		hashSize:      config.HashSize,
		purgeInterval: config.PurgeInterval,
	}
	for i := 0; i < len(URL64); i++ {
		app.handlerMap[URL64[i:i+1]] = newSessionHandler(config)
	}

	go app.purgeGoroutine()

	return app, nil
}

func (a *SessionStore) Create(data []byte) string {
	var id string
	err := errors.New("dummy")

	for err != nil {
		id = a.newId()
		err = a.getHandler(id).create(id, data)
	}

	return id
}

func (a *SessionStore) Update(id string, data []byte) error {
	return a.getHandler(id).update(id, data)
}

func (a *SessionStore) Delete(id string) error {
	return a.getHandler(id).delete(id)
}

func (a *SessionStore) Read(id string) ([]byte, error) {
	return a.getHandler(id).read(id)
}

func (a *SessionStore) Timestamp(id string) (*time.Time, error) {
	return a.getHandler(id).timestamp(id)
}

func (a *SessionStore) PurgeExpired() error {
	for key, handler := range a.handlerMap {
		if err := handler.purgeExpired(key); err != nil {
			return err
		}
	}

	return nil
}

func (a *SessionStore) purgeGoroutine() {
	for {
		<-time.After(a.purgeInterval)
		_ = a.PurgeExpired()
	}
}

func (a *SessionStore) getHandler(id string) *sessionHandler {
	return a.handlerMap[id[0:1]]
}

func (a *SessionStore) newId() string {
	var word string

	for i := 0; i < a.hashSize; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(rndSize))
		if err != nil {
			panic(err)
		}
		n := nBig.Int64()

		for j := 0; j < 10; j++ {
			m := n % 64
			word += URL64[m : m+1]
			n = n / 64
		}
	}

	return word
}
