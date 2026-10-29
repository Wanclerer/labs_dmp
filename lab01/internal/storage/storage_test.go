package storage

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSaveAndLookup(t *testing.T) {
	s := New()

	if err := s.Save("abc", "https://example.com"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok := s.Lookup("abc")
	if !ok {
		t.Fatal("Lookup: код не найден")
	}
	if got != "https://example.com" {
		t.Fatalf("Lookup = %q, ожидалось %q", got, "https://example.com")
	}
}

func TestSaveRejectsTakenCode(t *testing.T) {
	s := New()

	if err := s.Save("abc", "https://first.example"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := s.Save("abc", "https://second.example"); !errors.Is(err, ErrCodeTaken) {
		t.Fatalf("Save на занятый код вернул %v, ожидалась ErrCodeTaken", err)
	}

	if got, _ := s.Lookup("abc"); got != "https://first.example" {
		t.Fatalf("занятый код перезаписан: %q", got)
	}
}

func TestLookupMissing(t *testing.T) {
	if _, ok := New().Lookup("nope"); ok {
		t.Fatal("Lookup: несуществующий код найден")
	}
}

// TestConcurrentSaveSameCode проверяет атомарность вставки при параллельном доступе.
func TestConcurrentSaveSameCode(t *testing.T) {
	s := New()
	const workers = 50
	const code = "shared_code"

	var (
		wg           sync.WaitGroup
		successCount int64
		takenCount   int64
		otherErrors  int64
	)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			url := fmt.Sprintf("https://example.com/%d", id)
			err := s.Save(code, url)
			switch {
			case err == nil:
				atomic.AddInt64(&successCount, 1)
			case errors.Is(err, ErrCodeTaken):
				atomic.AddInt64(&takenCount, 1)
			default:
				atomic.AddInt64(&otherErrors, 1)
			}
		}(i)
	}
	wg.Wait()

	if successCount != 1 {
		t.Fatalf("успешных записей: %d, ожидалась ровно 1", successCount)
	}
	if takenCount != workers-1 {
		t.Fatalf("ошибок ErrCodeTaken: %d, ожидалось %d", takenCount, workers-1)
	}
	if otherErrors != 0 {
		t.Fatalf("неожиданных ошибок: %d", otherErrors)
	}

	winnerURL, ok := s.Lookup(code)
	if !ok || winnerURL == "" {
		t.Fatalf("Lookup вернул пустой результат для занятого кода: ok=%v, url=%q", ok, winnerURL)
	}
}