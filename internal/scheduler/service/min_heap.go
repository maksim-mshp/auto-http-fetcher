package service

import "auto-http-fetcher/internal/scheduler/domain"

type minHeap []*domain.QueueItem

func (h *minHeap) Len() int {
	return len(*h)
}

func (h *minHeap) Less(i, j int) bool {
	return (*h)[i].NextFetch.Before((*h)[j].NextFetch)
}

func (h *minHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
	(*h)[i].Index = i
	(*h)[j].Index = j
}

func (h *minHeap) Push(x any) {
	item := x.(*domain.QueueItem)
	item.Index = len(*h)
	*h = append(*h, item)
}

func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*h = old[:n-1]
	return item
}
