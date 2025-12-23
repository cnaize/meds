package types

import (
	"maps"
	"sync/atomic"

	"github.com/cnaize/meds/lib/util/get"
)

type CountryList struct {
	list atomic.Pointer[map[string]bool]
}

func NewCountryList() *CountryList {
	var l CountryList
	l.list.Store(get.Ptr(make(map[string]bool)))

	return &l
}

func (l *CountryList) GetAll() []string {
	list := *l.list.Load()
	coutries := make([]string, 0, len(list))
	for country := range maps.Keys(list) {
		coutries = append(coutries, country)
	}

	return coutries
}

func (l *CountryList) Lookup(country string) bool {
	return (*l.list.Load())[country]
}

func (l *CountryList) Upsert(coutries []string) error {
	list := maps.Clone(*l.list.Load())
	for _, country := range coutries {
		list[country] = true
	}

	l.list.Store(&list)

	return nil
}

func (l *CountryList) Remove(countries []string) error {
	list := maps.Clone(*l.list.Load())
	for _, country := range countries {
		delete(list, country)
	}

	l.list.Store(&list)

	return nil
}
