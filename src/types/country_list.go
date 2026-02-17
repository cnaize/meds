package types

import (
	"maps"
	"strings"
	"sync/atomic"
)

type CountryList struct {
	list atomic.Pointer[map[string]struct{}]
}

func NewCountryList() *CountryList {
	var l CountryList
	l.Store(new(make(map[string]struct{})))

	return &l
}

func (l *CountryList) Load() *map[string]struct{} {
	return l.list.Load()
}

func (l *CountryList) Store(list *map[string]struct{}) {
	l.list.Store(list)
}

func (l *CountryList) GetAll() []string {
	list := *l.Load()
	coutries := make([]string, 0, len(list))
	for country := range list {
		coutries = append(coutries, strings.ToLower(country))
	}

	return coutries
}

func (l *CountryList) Lookup(country string) bool {
	if _, ok := (*l.Load())[strings.ToLower(country)]; ok {
		return true
	}

	return false
}

func (l *CountryList) Upsert(coutries []string) error {
	list := maps.Clone(*l.Load())
	for _, country := range coutries {
		list[strings.ToLower(country)] = struct{}{}
	}

	l.Store(&list)

	return nil
}

func (l *CountryList) Remove(countries []string) error {
	list := maps.Clone(*l.Load())
	for _, country := range countries {
		delete(list, strings.ToLower(country))
	}

	l.Store(&list)

	return nil
}
