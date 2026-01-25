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
		handlerMap     map[string]*sessionHandler
		cryptoStrength int
		purgeInterval  time.Duration
	}

	SessionStoreConfig struct {
		SavePath       string
		CryptoStrength int
		ExpireInterval time.Duration
		PurgeInterval  time.Duration
	}
)

func NewSessionStoreConfig() SessionStoreConfig {
	workDir, _ := os.Getwd()
	return SessionStoreConfig{
		SavePath:       filepath.Join(workDir, "sessions"),
		CryptoStrength: 8,
		ExpireInterval: 365 * 24 * time.Hour,
		PurgeInterval:  24 * time.Hour,
	}
}

func NewSessionStore(cfg SessionStoreConfig) (*SessionStore, error) {
	fileInfo, err := os.Stat(cfg.SavePath)
	if errors.Is(err, fs.ErrNotExist) {
		if err = os.Mkdir(cfg.SavePath, 0700); err != nil {
			return nil, fmt.Errorf("mkdir: '%s' error: %w", cfg.SavePath, err)
		}
	} else if err != nil {
		slog.Error("creating session directory '%s' error: %w", cfg.SavePath, err)
		return nil, err
	} else if !fileInfo.IsDir() {
		return nil, errors.New(fmt.Sprintf("ids savePath is not a directory: %s", cfg.SavePath))
	}

	app := &SessionStore{
		handlerMap:     make(map[string]*sessionHandler),
		cryptoStrength: cfg.CryptoStrength,
		purgeInterval:  cfg.PurgeInterval,
	}
	for i := 0; i < len(URL64); i++ {
		app.handlerMap[URL64[i:i+1]] = newSessionHandler(cfg)
	}

	go app.purgeGoroutine()

	return app, nil
}

func (a *SessionStore) SID() string {
	var word string

	for i := 0; i < a.cryptoStrength; i++ {
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

func (a *SessionStore) Create(data []byte) string {
	var id string
	err := errors.New("dummy")

	for err != nil {
		id = a.SID()
		err = a.getHandler(id).create(id, data)
	}

	return id
}

func (a *SessionStore) Read(id string) ([]byte, error) {
	return a.getHandler(id).read(id)
}

func (a *SessionStore) Write(id string, data []byte) error {
	return a.getHandler(id).write(id, data)
}

func (a *SessionStore) Delete(id string) error {
	return a.getHandler(id).delete(id)
}

func (a *SessionStore) Timestamp(id string) (*time.Time, error) {
	return a.getHandler(id).timestamp(id)
}

func (a *SessionStore) Reset(id string) (string, error) {
	var newID string

	data, err := a.Read(id)
	if err != nil {
		return newID, err
	}

	err = a.Delete(id)
	if err != nil {
		return newID, err
	}

	return a.Create(data), nil
}

func (a *SessionStore) Purge() error {
	for key, handler := range a.handlerMap {
		if err := handler.purge(key); err != nil {
			return err
		}
	}

	return nil
}

func (a *SessionStore) purgeGoroutine() {
	for {
		<-time.After(a.purgeInterval)
		_ = a.Purge()
	}
}

func (a *SessionStore) getHandler(id string) *sessionHandler {
	return a.handlerMap[id[0:1]]
}
