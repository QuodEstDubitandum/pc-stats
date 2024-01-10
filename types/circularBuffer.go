package types

type CircularBuffer[T Data] struct {
    Data       	[]T
    Capacity   	int
    Head 		int
	Tail 		int
}

func NewCircularBuffer[T Data](capacity int) *CircularBuffer[T] {
	return &CircularBuffer[T]{
		Data: make([]T, capacity),
		Capacity: capacity,
		Head: 0,
		Tail: 0,
	}
}

func (cb *CircularBuffer[T]) Push(value T) {
	cb.Data[cb.Head] = value
	cb.Head = (cb.Head + 1) % cb.Capacity

	if cb.Head == cb.Tail {
		cb.Tail = (cb.Tail + 1) % cb.Capacity
	}
}
