package service

import (
	"auto-http-fetcher/internal/scheduler/domain"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"
	"container/heap"
	"time"
)

type PriorityQueue struct {
	heap  minHeap
	index map[int]*domain.QueueItem
}

func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		heap:  minHeap{},
		index: make(map[int]*domain.QueueItem),
	}
	heap.Init(&pq.heap)
	return pq
}

func (pq *PriorityQueue) Len() int {
	return pq.heap.Len()
}

func (pq *PriorityQueue) Add(id int, nextFetch time.Time, webhook *domainWebhook.Webhook) error {
	if _, exists := pq.index[id]; exists {
		return domain.AlreadyExistsError{ID: id}
	}

	item := &domain.QueueItem{
		ID:        id,
		NextFetch: nextFetch,
		Webhook:   webhook,
	}

	heap.Push(&pq.heap, item)
	pq.index[id] = item

	return nil
}

func (pq *PriorityQueue) Min() (*domain.QueueItem, bool) {
	if pq.heap.Len() == 0 {
		return nil, false
	}
	return pq.heap[0], true
}

func (pq *PriorityQueue) RemoveMin() (*domain.QueueItem, bool) {
	if pq.heap.Len() == 0 {
		return nil, false
	}

	item := heap.Pop(&pq.heap).(*domain.QueueItem)
	delete(pq.index, item.ID)

	return item, true
}

func (pq *PriorityQueue) Get(id int) (*domain.QueueItem, error) {
	item, ok := pq.index[id]
	if !ok {
		return nil, domain.NotFoundError{ID: id}
	}
	return item, nil
}

func (pq *PriorityQueue) Remove(id int) (*domain.QueueItem, error) {
	item, ok := pq.index[id]
	if !ok {
		return nil, domain.NotFoundError{ID: id}
	}

	removed := heap.Remove(&pq.heap, item.Index).(*domain.QueueItem)
	delete(pq.index, removed.ID)

	return removed, nil
}

func (pq *PriorityQueue) UpdateNextFetch(id int, nextFetch time.Time) error {
	item, ok := pq.index[id]
	if !ok {
		return domain.NotFoundError{ID: id}
	}

	item.NextFetch = nextFetch
	heap.Fix(&pq.heap, item.Index)

	return nil
}
