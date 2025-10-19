package container_test

import (
	"testing"

	"github.com/jeffersonbrasilino/gomes/container"
)

func TestGenericContainer_Set(t *testing.T) {
	t.Parallel()
	c := container.NewGenericContainer[string, int]()
	key := "foo"
	item := 42

	t.Run("should set item successfully", func(t *testing.T) {
		err := c.Set(key, item)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should fail to set item if key exists", func(t *testing.T) {
		_ = c.Set(key, item)
		err := c.Set(key, item)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestGenericContainer_Has(t *testing.T) {
	t.Parallel()
	c := container.NewGenericContainer[string, int]()
	key := "foo"
	item := 42

	t.Run("should return false if key does not exist", func(t *testing.T) {
		if c.Has(key) {
			t.Errorf("expected Has to be false")
		}
	})

	t.Run("should return true if key exists", func(t *testing.T) {
		_ = c.Set(key, item)
		if !c.Has(key) {
			t.Errorf("expected Has to be true")
		}
	})
}

func TestGenericContainer_Replace(t *testing.T) {
	t.Parallel()
	c := container.NewGenericContainer[string, int]()
	key := "foo"
	item := 42
	newItem := 99

	t.Run("should fail to replace if key does not exist", func(t *testing.T) {
		err := c.Replace(key, newItem)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should replace item if key exists", func(t *testing.T) {
		_ = c.Set(key, item)
		err := c.Replace(key, newItem)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		val, _ := c.Get(key)
		if val != newItem {
			t.Errorf("expected value to be replaced")
		}
	})
}

func TestGenericContainer_Get(t *testing.T) {
	t.Parallel()
	c := container.NewGenericContainer[string, int]()
	key := "foo"
	item := 42

	t.Run("should fail to get if key does not exist", func(t *testing.T) {
		_, err := c.Get(key)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should get item if key exists", func(t *testing.T) {
		_ = c.Set(key, item)
		val, err := c.Get(key)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if val != item {
			t.Errorf("expected value to be %v, got %v", item, val)
		}
	})
}

func TestGenericContainer_GetAll(t *testing.T) {
	t.Parallel()
	c := container.NewGenericContainer[string, int]()
	key := "foo"
	item := 42
	_ = c.Set(key, item)
	all := c.GetAll()
	if len(all) != 1 {
		t.Errorf("expected 1 item, got %d", len(all))
	}
	if all[key] != item {
		t.Errorf("expected value to be %v, got %v", item, all[key])
	}
}

func TestGenericContainer_Remove(t *testing.T) {
	t.Parallel()
	c := container.NewGenericContainer[string, int]()
	key := "foo"
	item := 42

	t.Run("should fail to remove if key does not exist", func(t *testing.T) {
		err := c.Remove(key)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should remove item if key exists", func(t *testing.T) {
		_ = c.Set(key, item)
		err := c.Remove(key)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if c.Has(key) {
			t.Errorf("expected key to be removed")
		}
	})
}
