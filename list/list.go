package list

type Element struct {
	prev, next *Element

	l *List

	Value interface{}
}

func (e *Element) Next() *Element {
	if e.l != nil && e.next == &e.l.root {
		return nil
	}
	return e.next
}

func (e *Element) Prev() *Element {
	if e.l != nil && e.prev == &e.l.root {
		return nil
	}
	return e.prev
}

type List struct {
	root Element

	len	int
}

func (l *List) init() *List {
	l.root.prev = &l.root
	l.root.next = &l.root
	l.len = 0

	return l
}

func New() *List {
	return new(List).init()
}

func (l *List) Front() *Element {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

func (l *List) Back() *Element {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

func (l *List) Len() int {
	return l.len
}

func (l *List) PushFront(v interface{}) *Element {
	if l.root.next == nil {
		l.init()
	}
	return l.insertValue(v, &l.root)
}

func (l *List) PushBack(v interface{}) *Element {
	if l.root.next == nil {
		l.init()
	}
	return l.insertValue(v, l.root.prev)
}

func (l *List) Remove(e *Element) interface{} {
	if e.l == l {
		l.remove(e)
	}
	return e.Value
}

func (l *List) remove(e *Element) interface{} {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil
	e.prev = nil
	e.l = nil
	l.len--
	return e.Value
}

func (l *List) insertValue(v interface{}, mark *Element) *Element {
	return l.insert(&Element{Value: v}, mark)
}

func (l *List) insert(e, mark *Element) *Element {
	e.l = l
	mark.next.prev = e
	e.prev = mark
	e.next = mark.next
	mark.next = e
	l.len++
	return e
}

func (l *List) MoveToFront(e *Element) {
	if e.l != l || e.l.root.next == e {
		return
	}
	l.remove(e)
	l.insert(e, &l.root)
}

func (l *List) MoveToBack(e *Element) {
	if e.l != l || e.l.root.prev == e {
		return
	}
	l.remove(e)
	l.insert(e, l.root.prev)
}

func (l *List) InsertAfter(v interface{}, mark *Element) *Element {
	if mark.l != l {
		return nil
	}
	return l.insertValue(v, mark)
}

func (l *List) InsertBefore(v interface{}, mark *Element) *Element {
	if mark.l != l {
		return nil
	}
	return l.insertValue(v, mark.prev)
}

func (l *List) MoveAfter(e, mark *Element) {
	if e.l != l || mark.l != l || e == mark {
		return
	}
	l.remove(e)
	l.insert(e, mark)
}

func (l *List) MoveBefore(e, mark *Element) {
	if e.l != l || mark.l != l || e == mark {
		return
	}
	l.remove(e)
	l.insert(e, mark.prev)	
}

func (l *List) PushBackList(other *List) {
	for i, e := other.Len(), other.Front(); i > 0 ; i, e = i - 1, e.Next() {
		l.PushBack(e.Value)
	}
}

func (l *List) PushFrontList(other *List) {
	for i , e := other.Len(), other.Back(); i > 0 ; i, e = i - 1, e.Prev() {
		l.PushFront(e.Value)
	}
}
