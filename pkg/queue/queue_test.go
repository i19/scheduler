package queue

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test_Queue(t *testing.T) {
	priority := []string{"high", "mid", "low"}
	count := 10

	q := NewPriorityQueue(priority, 2)

	for _, p := range priority {
		for i := 1; i < count; i++ {
			q.Push(p, fmt.Sprintf("%s_%d", p, i))
		}
	}

	out := q.Pop()
	var result []string
	for {
		select {
		case x := <-out:
			result = append(result, x.(string))
		case <-time.After(time.Second):
			println(strings.Join(result, ","))
			return
		}
	}

	//for x := range q.Pop() {
	//	println(x.(string))
	//}
}
