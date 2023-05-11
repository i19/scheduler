package queue

import (
	"math"
)

type PriorityQueue struct {
	order   []string
	dataMap map[string]chan interface{}
	counter []int
	factor  int
}

func NewPriorityQueue(order []string, factor int) *PriorityQueue {
	pq := &PriorityQueue{
		order:   order,
		dataMap: make(map[string]chan interface{}),
		counter: make([]int, len(order)),
		factor:  factor,
	}

	for _, o := range order {
		pq.dataMap[o] = make(chan interface{}, 1000)
	}
	return pq
}

func (pq *PriorityQueue) Push(priority string, data interface{}) {
	pq.dataMap[priority] <- data
}

func (pq *PriorityQueue) Pop() chan interface{} {
	ch := make(chan interface{})
	go func(ch chan interface{}) {
		orderBase := make([]int, len(pq.order))
		for i := range pq.order {
			orderBase[i] = int(math.Pow(float64(pq.factor), float64(len(pq.order)-i-1)))
		}
		for {
			send := false
		LOOP:
			for i, order := range pq.order {
				if pq.counter[i] < orderBase[i] {
					select {
					case v := <-pq.dataMap[order]:
						pq.counter[i]++
						ch <- v
						send = true
						break LOOP
					default:
						send = false
					}
				}
			}
			if !send {
				pq.counter = make([]int, len(pq.order))
			}
		}
	}(ch)
	return ch
}
