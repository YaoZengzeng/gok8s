package priorityqueue

import (
	"container/heap"
)

type Item struct {
	value		string
	priority 	int
	index		int
}

type PriorityQueue []*Item

func (q PriorityQueue) Less(i, j int) bool {
	return q[i].priority > q[j].priority
}

func (q PriorityQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q PriorityQueue) Len() int {
	return len(q)
}

func (q *PriorityQueue) Push(x interface{}) {
	item := x.(*Item)
	item.index = len(*q)
	*q = append(*q, item)	
}

func (q *PriorityQueue) Pop() interface{} {
	item := (*q)[len(*q) - 1]
	item.index = -1
	*q = (*q)[: len(*q) - 1]
	return item
}

func (q *PriorityQueue) Update(item *Item, value string, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(q, item.index)
}
