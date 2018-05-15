package priorityqueue

import (
	"container/heap"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	m := map[string]int{
		"abc": 3,
		"def": 1,
		"ghi": 10,
	}

	var q PriorityQueue
	index := 0
	for value, priority := range m {
		q = append(q, &Item{
			value:    value,
			priority: priority,
			index:    index,
		})
		index++
	}

	heap.Init(&q)

	item := &Item{
		value:    "jkl",
		priority: 5,
	}

	heap.Push(&q, item)

	q.Update(item, "xyz", 20)

	cases := []struct {
		value    string
		priority int
	}{
		{"xyz", 20},
		{"ghi", 10},
		{"abc", 3},
		{"def", 1},
	}
	for _, c := range cases {
		item = heap.Pop(&q).(*Item)
		if c.value != item.value || c.priority != item.priority {
			t.Fatalf("Unexpected item, expected: %s/%d, real: %s/%d", c.value, c.priority, item.value, item.priority)
		}
	}

	if len(q) != 0 {
		t.Fatalf("PriorityQueue should be empty finally")
	}
}
