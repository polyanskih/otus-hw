package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)
		require.False(t, c.Set("aaa", 100))
		require.False(t, c.Set("bbb", 200))
		require.False(t, c.Set("ccc", 300))
		require.False(t, c.Set("ddd", 400))

		val, ok := c.Get("aaa")
		require.False(t, ok)
		require.Nil(t, val)
		for key, expected := range map[Key]int{"bbb": 200, "ccc": 300, "ddd": 400} {
			val, ok := c.Get(key)
			require.True(t, ok)
			require.Equal(t, expected, val)
		}
	})
}

func TestCacheRecentlyUsed(t *testing.T) {
	t.Run("get updates order", func(t *testing.T) {
		c := NewCache(3)
		c.Set("aaa", 100)
		c.Set("bbb", 200)
		c.Set("ccc", 300)

		for i := 0; i < 2; i++ {
			val, ok := c.Get("aaa")
			require.True(t, ok)
			require.Equal(t, 100, val)
		}
		c.Set("ddd", 400)

		val, ok := c.Get("bbb")
		require.False(t, ok)
		require.Nil(t, val)
		for key, expected := range map[Key]int{"aaa": 100, "ccc": 300, "ddd": 400} {
			val, ok := c.Get(key)
			require.True(t, ok)
			require.Equal(t, expected, val)
		}
	})

	t.Run("set updates order", func(t *testing.T) {
		c := NewCache(3)
		c.Set("aaa", 100)
		c.Set("bbb", 200)
		c.Set("ccc", 300)

		require.True(t, c.Set("aaa", 500))
		require.True(t, c.Set("aaa", 600))
		c.Set("ddd", 400)

		val, ok := c.Get("bbb")
		require.False(t, ok)
		require.Nil(t, val)
		for key, expected := range map[Key]int{"aaa": 600, "ccc": 300, "ddd": 400} {
			val, ok := c.Get(key)
			require.True(t, ok)
			require.Equal(t, expected, val)
		}
	})

	t.Run("mixed access and reinsertion", func(t *testing.T) {
		c := NewCache(3)
		c.Set("aaa", 100)
		c.Set("bbb", 200)
		c.Set("ccc", 300)
		c.Get("aaa")
		require.True(t, c.Set("bbb", 500))
		c.Get("ccc")

		val, ok := c.Get("missing")
		require.False(t, ok)
		require.Nil(t, val)
		c.Set("ddd", 400)

		val, ok = c.Get("aaa")
		require.False(t, ok)
		require.Nil(t, val)
		require.False(t, c.Set("aaa", 600))
		val, ok = c.Get("bbb")
		require.False(t, ok)
		require.Nil(t, val)
		for key, expected := range map[Key]int{"aaa": 600, "ccc": 300, "ddd": 400} {
			val, ok := c.Get(key)
			require.True(t, ok)
			require.Equal(t, expected, val)
		}
	})
}

func TestCacheCapacity(t *testing.T) {
	t.Run("zero capacity", func(t *testing.T) {
		c := NewCache(0)
		for i := 0; i < 2; i++ {
			require.False(t, c.Set("aaa", 100))
			val, ok := c.Get("aaa")
			require.False(t, ok)
			require.Nil(t, val)
		}
	})

	t.Run("single item", func(t *testing.T) {
		c := NewCache(1)
		require.False(t, c.Set("aaa", 100))
		require.True(t, c.Set("aaa", 200))
		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 200, val)

		require.False(t, c.Set("bbb", 300))
		val, ok = c.Get("aaa")
		require.False(t, ok)
		require.Nil(t, val)
		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 300, val)
	})
}

func TestCacheClear(t *testing.T) {
	c := NewCache(2)
	c.Clear()
	c.Set("aaa", 100)
	c.Set("bbb", 200)
	c.Clear()
	c.Clear()
	for _, key := range []Key{"aaa", "bbb"} {
		val, ok := c.Get(key)
		require.False(t, ok)
		require.Nil(t, val)
	}

	require.False(t, c.Set("aaa", 300))
	require.False(t, c.Set("bbb", 400))
	require.False(t, c.Set("ccc", 500))
	val, ok := c.Get("aaa")
	require.False(t, ok)
	require.Nil(t, val)
	for key, expected := range map[Key]int{"bbb": 400, "ccc": 500} {
		val, ok := c.Get(key)
		require.True(t, ok)
		require.Equal(t, expected, val)
	}
}

func TestCacheNilValue(t *testing.T) {
	c := NewCache(1)
	require.False(t, c.Set("", nil))
	val, ok := c.Get("")
	require.True(t, ok)
	require.Nil(t, val)

	require.True(t, c.Set("", "value"))
	val, ok = c.Get("")
	require.True(t, ok)
	require.Equal(t, "value", val)
}

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
	val, ok := c.Get("999999")
	require.True(t, ok)
	require.Equal(t, 999999, val)
}

func TestCacheConcurrentClear(t *testing.T) {
	c := NewCache(10)
	start := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(5)

	for worker := 0; worker < 4; worker++ {
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < 1_000; i++ {
				value := i % 20
				key := Key(strconv.Itoa(value))
				c.Set(key, value)
				val, ok := c.Get(key)
				if ok && val != value {
					t.Errorf("Get(%q) = %v, want %d", key, val, value)
					return
				}
				if !ok && val != nil {
					t.Errorf("Get(%q) = %v on cache miss, want nil", key, val)
					return
				}
			}
		}()
	}

	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 1_000; i++ {
			c.Clear()
		}
	}()

	close(start)
	wg.Wait()

	c.Clear()
	for i := 0; i < 20; i++ {
		val, ok := c.Get(Key(strconv.Itoa(i)))
		require.False(t, ok)
		require.Nil(t, val)
	}
	for i := 0; i <= 10; i++ {
		require.False(t, c.Set(Key(strconv.Itoa(i)), i))
	}
	val, ok := c.Get("0")
	require.False(t, ok)
	require.Nil(t, val)
	for i := 1; i <= 10; i++ {
		val, ok := c.Get(Key(strconv.Itoa(i)))
		require.True(t, ok)
		require.Equal(t, i, val)
	}
}
