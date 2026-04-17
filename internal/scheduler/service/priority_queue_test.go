package service

import (
	"auto-http-fetcher/internal/scheduler/domain"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"
	"errors"
	"testing"
	"time"
)

func TestPriorityQueue_EmptyQueue(t *testing.T) {
	pq := NewPriorityQueue()

	if pq.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", pq.Len())
	}

	if item, ok := pq.Min(); ok || item != nil {
		t.Fatalf("Min() = (%v, %v), want (nil, false)", item, ok)
	}

	if item, ok := pq.RemoveMin(); ok || item != nil {
		t.Fatalf("RemoveMin() = (%v, %v), want (nil, false)", item, ok)
	}

	if _, err := pq.Get(1); !errors.As(err, &domain.NotFoundError{}) {
		t.Fatalf("Get() error = %v, want %T", err, domain.NotFoundError{})
	}

	if _, err := pq.Remove(1); !errors.As(err, &domain.NotFoundError{}) {
		t.Fatalf("Remove() error = %v, want %T", err, domain.NotFoundError{})
	}

	if err := pq.UpdateNextFetch(1, time.Now()); !errors.As(err, &domain.NotFoundError{}) {
		t.Fatalf("UpdateNextFetch() error = %v, want %T", err, domain.NotFoundError{})
	}
}

func TestPriorityQueue_AddDuplicateID(t *testing.T) {
	pq := NewPriorityQueue()
	now := time.Date(2026, time.April, 17, 12, 0, 0, 0, time.UTC)

	if err := pq.Add(7, now, testWebhook(7)); err != nil {
		t.Fatalf("Add() first call error = %v", err)
	}

	err := pq.Add(7, now.Add(time.Minute), testWebhook(70))
	if !errors.As(err, &domain.AlreadyExistsError{}) {
		t.Fatalf("Add() duplicate error = %v, want %T", err, domain.AlreadyExistsError{})
	}

	if pq.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", pq.Len())
	}

	item, ok := pq.Min()
	if !ok {
		t.Fatal("Min() ok = false, want true")
	}
	if item.ID != 7 {
		t.Fatalf("Min().ID = %d, want 7", item.ID)
	}
	if !item.NextFetch.Equal(now) {
		t.Fatalf("Min().NextFetch = %v, want %v", item.NextFetch, now)
	}
}

func TestPriorityQueue_RemoveMinReturnsItemsInAscendingOrder(t *testing.T) {
	pq := NewPriorityQueue()
	base := time.Date(2026, time.April, 17, 12, 0, 0, 0, time.UTC)

	items := []struct {
		id   int
		next time.Time
	}{
		{id: 10, next: base.Add(3 * time.Minute)},
		{id: 20, next: base.Add(1 * time.Minute)},
		{id: 30, next: base.Add(2 * time.Minute)},
	}

	for _, tc := range items {
		if err := pq.Add(tc.id, tc.next, testWebhook(tc.id)); err != nil {
			t.Fatalf("Add(%d) error = %v", tc.id, err)
		}
	}

	wantOrder := []int{20, 30, 10}
	for i, wantID := range wantOrder {
		item, ok := pq.RemoveMin()
		if !ok {
			t.Fatalf("RemoveMin() on step %d returned ok = false", i)
		}
		if item.ID != wantID {
			t.Fatalf("RemoveMin() on step %d returned ID = %d, want %d", i, item.ID, wantID)
		}
		if item.Index != -1 {
			t.Fatalf("removed item Index = %d, want -1", item.Index)
		}
		if _, err := pq.Get(wantID); !errors.As(err, &domain.NotFoundError{}) {
			t.Fatalf("Get(%d) after RemoveMin() error = %v, want %T", wantID, err, domain.NotFoundError{})
		}
	}

	if pq.Len() != 0 {
		t.Fatalf("Len() after draining = %d, want 0", pq.Len())
	}
}

func TestPriorityQueue_RemoveDeletesExactItemAndKeepsHeapValid(t *testing.T) {
	pq := NewPriorityQueue()
	base := time.Date(2026, time.April, 17, 12, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		id   int
		next time.Time
	}{
		{id: 1, next: base.Add(4 * time.Minute)},
		{id: 2, next: base.Add(1 * time.Minute)},
		{id: 3, next: base.Add(3 * time.Minute)},
		{id: 4, next: base.Add(2 * time.Minute)},
	} {
		if err := pq.Add(tc.id, tc.next, testWebhook(tc.id)); err != nil {
			t.Fatalf("Add(%d) error = %v", tc.id, err)
		}
	}

	removed, err := pq.Remove(3)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if removed.ID != 3 {
		t.Fatalf("Remove() ID = %d, want 3", removed.ID)
	}
	if removed.Index != -1 {
		t.Fatalf("removed item Index = %d, want -1", removed.Index)
	}

	if _, err := pq.Get(3); !errors.As(err, &domain.NotFoundError{}) {
		t.Fatalf("Get(3) after Remove() error = %v, want %T", err, domain.NotFoundError{})
	}

	minItem, ok := pq.Min()
	if !ok {
		t.Fatal("Min() ok = false, want true")
	}
	if minItem.ID != 2 {
		t.Fatalf("Min().ID = %d, want 2", minItem.ID)
	}

	wantOrder := []int{2, 4, 1}
	for i, wantID := range wantOrder {
		item, ok := pq.RemoveMin()
		if !ok {
			t.Fatalf("RemoveMin() on step %d returned ok = false", i)
		}
		if item.ID != wantID {
			t.Fatalf("RemoveMin() on step %d returned ID = %d, want %d", i, item.ID, wantID)
		}
	}
}

func TestPriorityQueue_UpdateNextFetchReordersItems(t *testing.T) {
	pq := NewPriorityQueue()
	base := time.Date(2026, time.April, 17, 12, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		id   int
		next time.Time
	}{
		{id: 1, next: base.Add(1 * time.Minute)},
		{id: 2, next: base.Add(2 * time.Minute)},
		{id: 3, next: base.Add(3 * time.Minute)},
	} {
		if err := pq.Add(tc.id, tc.next, testWebhook(tc.id)); err != nil {
			t.Fatalf("Add(%d) error = %v", tc.id, err)
		}
	}

	if err := pq.UpdateNextFetch(3, base.Add(30*time.Second)); err != nil {
		t.Fatalf("UpdateNextFetch() earlier error = %v", err)
	}

	minItem, ok := pq.Min()
	if !ok {
		t.Fatal("Min() after earlier update ok = false, want true")
	}
	if minItem.ID != 3 {
		t.Fatalf("Min().ID after earlier update = %d, want 3", minItem.ID)
	}
	if !minItem.NextFetch.Equal(base.Add(30 * time.Second)) {
		t.Fatalf("Min().NextFetch after earlier update = %v, want %v", minItem.NextFetch, base.Add(30*time.Second))
	}

	if err := pq.UpdateNextFetch(3, base.Add(10*time.Minute)); err != nil {
		t.Fatalf("UpdateNextFetch() later error = %v", err)
	}

	minItem, ok = pq.Min()
	if !ok {
		t.Fatal("Min() after later update ok = false, want true")
	}
	if minItem.ID != 1 {
		t.Fatalf("Min().ID after later update = %d, want 1", minItem.ID)
	}

	item, err := pq.Get(3)
	if err != nil {
		t.Fatalf("Get(3) error = %v", err)
	}
	if !item.NextFetch.Equal(base.Add(10 * time.Minute)) {
		t.Fatalf("Get(3).NextFetch = %v, want %v", item.NextFetch, base.Add(10*time.Minute))
	}
}

func testWebhook(id int) *domainWebhook.Webhook {
	return &domainWebhook.Webhook{ID: id}
}
