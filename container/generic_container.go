// Package container provides a generic container implementation for managing
// key-value pairs with thread-safe operations. It supports setting, getting,
// replacing, and removing items, with error handling for duplicate keys and
// missing items.
package container

import (
	"fmt"
	"maps"
	"sync"
)

type (
	genericContainer[K comparable, T any] struct {
		mu        sync.RWMutex
		container map[K]T
	}
	// Container defines the interface for a generic key-value container.
	// It provides methods to manage items with error handling.
	Container[K comparable, T any] interface {
		// Set adds a new item to the container. Returns an error if the key already exists.
		Set(key K, item T) error
		// Has checks if a key exists in the container.
		Has(key K) bool
		// Replace updates an existing item. Returns an error if the key does not exist.
		Replace(key K, item T) error
		// Get retrieves an item by key. Returns an error if the key is not found.
		Get(key K) (T, error)
		// GetAll returns a copy of all items in the container.
		GetAll() map[K]T
		// Remove deletes an item by key. Returns an error if the key is not found.
		Remove(key K) error
	}
)

// NewGenericContainer creates a new instance of a generic container.
// It initializes an empty map for storing key-value pairs.
func NewGenericContainer[K comparable, T any]() *genericContainer[K, T] {
	return &genericContainer[K, T]{container: make(map[K]T)}
}

func (c *genericContainer[K, T]) Set(key K, item T) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, found := c.container[key]; found {
		return fmt.Errorf("%v already exists", key)
	}
	c.container[key] = item
	return nil
}

func (c *genericContainer[K, T]) Has(key K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, found := c.container[key]
	return found
}

func (c *genericContainer[K, T]) Replace(key K, item T) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, found := c.container[key]
	if !found {
		return fmt.Errorf("cannot find item %v", key)
	}
	c.container[key] = item
	return nil
}

func (c *genericContainer[K, T]) Get(key K) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var item T
	item, found := c.container[key]
	if !found {
		return item, fmt.Errorf("cannot find item %v", key)
	}
	return item, nil
}

func (c *genericContainer[K, T]) GetAll() map[K]T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return maps.Clone(c.container)
}

func (c *genericContainer[K, T]) Remove(key K) error {

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, found := c.container[key]; !found {
		return fmt.Errorf("cannot find item %v", key)
	}

	delete(c.container, key)
	return nil
}
