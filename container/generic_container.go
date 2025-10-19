package container

import "fmt"

type (
	genericContainer[K comparable, T any] struct {
		container map[K]T
	}
	Container[K comparable, T any] interface {
		Set(key K, item T) error
		Has(key K) bool
		Replace(key K, item T) error
		Get(key K) (T, error)
		GetAll() map[K]T
		Remove(key K) error
	}
)

func NewGenericContainer[K comparable, T any]() *genericContainer[K, T] {
	return &genericContainer[K, T]{container: make(map[K]T)}
}

func (c genericContainer[K, T]) Set(key K, item T) error {
	if c.Has(key) {
		return fmt.Errorf("%v already exists", key)
	}
	c.container[key] = item
	return nil
}

func (c genericContainer[K, T]) Has(key K) bool {
	_, found := c.container[key]
	return found
}

func (c genericContainer[K, T]) Replace(key K, item T) error {
	_, found := c.container[key]
	if !found {
		return fmt.Errorf("cannot find item %v", key)
	}
	c.container[key] = item
	return nil
}

func (c genericContainer[K, T]) Get(key K) (T, error) {
	var item T
	item, found := c.container[key]
	if !found {
		return item, fmt.Errorf("cannot find item %v", key)
	}
	return item, nil
}

func (c *genericContainer[K, T]) GetAll() map[K]T {
	return c.container
}

func (c *genericContainer[K, T]) Remove(key K) error {

	if !c.Has(key) {
		return fmt.Errorf("cannot find item %v", key)
	}

	delete(c.container, key)
	return nil
}
