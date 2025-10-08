package list

type List interface {
	Lookup(item string) bool
	Upsert(items []string) error
	Remove(items []string) error
}
