package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}

func TestListPush(t *testing.T) {
	t.Run("push front", func(t *testing.T) {
		l := NewList()
		first := l.PushFront(10)
		requireListItems(t, l, first)

		second := l.PushFront("value")
		requireListItems(t, l, second, first)
		require.Equal(t, 10, first.Value)
		require.Equal(t, "value", second.Value)
	})

	t.Run("push back", func(t *testing.T) {
		l := NewList()
		first := l.PushBack(nil)
		requireListItems(t, l, first)

		second := l.PushBack(20)
		requireListItems(t, l, first, second)
		require.Nil(t, first.Value)
		require.Equal(t, 20, second.Value)
	})
}

func TestListRemove(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected []int
	}{
		{name: "front", index: 0, expected: []int{1, 2}},
		{name: "middle", index: 1, expected: []int{0, 2}},
		{name: "back", index: 2, expected: []int{0, 1}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := NewList()
			items := []*ListItem{l.PushBack(10), l.PushBack(20), l.PushBack(30)}
			l.Remove(items[tc.index])
			requireListItems(t, l, items[tc.expected[0]], items[tc.expected[1]])
		})
	}

	t.Run("only item and reuse", func(t *testing.T) {
		l := NewList()
		item := l.PushFront(10)
		l.Remove(item)
		requireListItems(t, l)

		item = l.PushBack(20)
		requireListItems(t, l, item)
		l.Remove(item)
		requireListItems(t, l)

		item = l.PushFront(30)
		requireListItems(t, l, item)
	})
}

func TestListMoveToFront(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected []int
	}{
		{name: "front", index: 0, expected: []int{0, 1, 2}},
		{name: "middle", index: 1, expected: []int{1, 0, 2}},
		{name: "back", index: 2, expected: []int{2, 0, 1}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := NewList()
			items := []*ListItem{l.PushBack(10), l.PushBack(20), l.PushBack(30)}
			l.MoveToFront(items[tc.index])
			requireListItems(t, l, items[tc.expected[0]], items[tc.expected[1]], items[tc.expected[2]])

			l.MoveToFront(items[tc.index])
			requireListItems(t, l, items[tc.expected[0]], items[tc.expected[1]], items[tc.expected[2]])
			require.Equal(t, 10, items[0].Value)
			require.Equal(t, 20, items[1].Value)
			require.Equal(t, 30, items[2].Value)
		})
	}

	t.Run("only item", func(t *testing.T) {
		l := NewList()
		item := l.PushBack(10)
		l.MoveToFront(item)
		requireListItems(t, l, item)
	})
}

func requireListItems(t *testing.T, l List, items ...*ListItem) {
	t.Helper()
	total := len(items)
	require.Equal(t, total, l.Len())
	if total == 0 {
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
		return
	}

	require.Same(t, items[0], l.Front())
	require.Same(t, items[total-1], l.Back())
	require.Nil(t, l.Front().Prev)
	require.Nil(t, l.Back().Next)
	for i := 1; i < total; i++ {
		require.Same(t, items[i], items[i-1].Next)
		require.Same(t, items[i-1], items[i].Prev)
	}
}
