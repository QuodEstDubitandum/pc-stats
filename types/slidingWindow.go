package types

type Data interface {
    float64 | uint64
}

type SlidingWindow[T Data] struct {
    Data     []T
    Capacity int
}

func NewSlidingWindow[T Data](capacity int) *SlidingWindow[T]{
	return &SlidingWindow[T]{
		Data: make([]T, 0, capacity),
		Capacity: capacity,
	}
}

func (sw *SlidingWindow[T]) Push(value T){
	if len(sw.Data) == sw.Capacity {
		sw.Data = sw.Data[1:]
	}

	sw.Data = append(sw.Data, value)
}