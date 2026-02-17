package types

import (
	"maps"
	"sync/atomic"
)

type MapList[T comparable] struct {
	list atomic.Pointer[map[T]struct{}]
}

func NewMapList[T comparable]() *MapList[T] {
	var l MapList[T]
	l.Store(new(make(map[T]struct{})))

	return &l
}

func (l *MapList[T]) Load() *map[T]struct{} {
	return l.list.Load()
}

func (l *MapList[T]) Store(list *map[T]struct{}) {
	l.list.Store(list)
}

func (l *MapList[T]) GetAll() []T {
	list := *l.Load()
	items := make([]T, 0, len(list))
	for item := range list {
		items = append(items, item)
	}

	return items
}

func (l *MapList[T]) Lookup(item T) bool {
	if _, ok := (*l.Load())[item]; ok {
		return true
	}

	return false
}

func (l *MapList[T]) Upsert(items []T) error {
	list := maps.Clone(*l.Load())
	for _, item := range items {
		list[item] = struct{}{}
	}

	l.Store(&list)

	return nil
}

func (l *MapList[T]) Remove(items []T) error {
	list := maps.Clone(*l.Load())
	for _, item := range items {
		delete(list, item)
	}

	l.Store(&list)

	return nil
}
