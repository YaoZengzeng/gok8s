package heap

type Interface interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
	Push(x interface{})
	Pop() interface{}
}

func Init(h Interface) {
	n := h.Len()
	for i := (n + 1) / 2; i >= 0; i-- {
		down(h, i, n)
	}
}

func Pop(h Interface) interface{} {
	n := h.Len() - 1
	h.Swap(0, n)
	down(h, 0, n)
	return h.Pop()
}

func Push(h Interface, x interface{}) {
	h.Push(x)
	up(h, h.Len()-1)
}

func Remove(h Interface, i int) interface{} {
	n := h.Len() - 1
	h.Swap(i, n)
	down(h, i, n)
	return h.Pop()
}

func Fix(h Interface, i int) {
	if !down(h, i, h.Len()) {
		up(h, i)
	}
}

func up(h Interface, i int) {
	for i > 0 {
		if h.Less(i, (i-1)/2) {
			h.Swap(i, (i-1)/2)
			i = (i - 1) / 2
		} else {
			break
		}
	}
}

func down(h Interface, i, n int) bool {
	i0 := i
	for {
		j := i0*2 + 1
		if j >= n {
			break
		}
		j2 := i0*2 + 2
		if j2 < n && h.Less(j2, j) {
			j = j2
		}
		if h.Less(j, i0) {
			h.Swap(j, i0)
			i0 = j
		} else {
			break
		}
	}
	return i0 != i
}
