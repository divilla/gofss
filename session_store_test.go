package fss

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	ss            *SessionStore
	data          []byte
	ids           []string
	readCounter   int64
	updateCounter int64
	deleteCounter int64
)

func BenchmarkSetup(b *testing.B) {
	var err error

	config := NewSessionStoreConfig()
	config.ExpireInterval = time.Minute
	ss, err = NewSessionStore(config)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		data = append(data, []byte(URL64)...)
	}

	b.SkipNow()
}

func BenchmarkSessionCreate(b *testing.B) {
	var wg sync.WaitGroup
	ch := make(chan string)

	go func() {
		for {
			id := <-ch
			ids = append(ids, id)
			wg.Done()
		}
	}()

	for n := 0; n < b.N; n++ {
		wg.Add(1)
		go func() {
			id := ss.Create(data)
			ch <- id
		}()
	}
	wg.Wait()
}

func BenchmarkSessionRead(b *testing.B) {
	var wg sync.WaitGroup
	for n := 0; n < b.N; n++ {
		if int(atomic.LoadInt64(&readCounter)) >= len(ids) {
			atomic.StoreInt64(&readCounter, 0)
		}

		wg.Add(1)
		go func(i int) {
			_, err := ss.Read(ids[i])
			if err != nil {
				b.Error(err)
			}
			wg.Done()
		}(int(atomic.LoadInt64(&readCounter)))

		atomic.AddInt64(&readCounter, 1)
	}

	wg.Wait()
}

func BenchmarkSessionUpdate(b *testing.B) {
	var wg sync.WaitGroup
	for n := 0; n < b.N; n++ {
		if int(atomic.LoadInt64(&updateCounter)) >= len(ids) {
			atomic.StoreInt64(&updateCounter, 0)
		}

		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			err := ss.Update(ids[j], data)
			if err != nil {
				b.Error(err)
			}
		}(int(atomic.LoadInt64(&updateCounter)))

		atomic.AddInt64(&updateCounter, 1)
	}

	wg.Wait()
}

func BenchmarkSessionDelete(b *testing.B) {
	var wg sync.WaitGroup
	for n := 0; n < b.N; n++ {
		if int(atomic.LoadInt64(&deleteCounter)) >= len(ids) {
			break
		}

		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			err := ss.Delete(ids[j])
			if err != nil {
				b.Error(err)
			}
		}(int(atomic.LoadInt64(&deleteCounter)))

		atomic.AddInt64(&deleteCounter, 1)
	}

	wg.Wait()
}

func BenchmarkClean(b *testing.B) {
	b.SkipNow()
	_ = ss.Purge()
}

func BenchmarkCleanup(b *testing.B) {
	for _, id := range ids {
		_ = ss.Delete(id)
	}

	b.SkipNow()
}
