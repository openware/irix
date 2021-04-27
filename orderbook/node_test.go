package orderbook

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestPushPop(t *testing.T) {
	s := newStack()
	var nSlice []*node
	for i := 0; i < 100; i++ {
		nSlice = append(nSlice, s.Pop())
	}

	if s.getCount() != 0 {
		t.Fatalf("incorrect stack count expected %v but received %v", 0, s.getCount())
	}

	for i := 0; i < 100; i++ {
		s.Push(nSlice[i], getNow())
	}

	if s.getCount() != 100 {
		t.Fatalf("incorrect stack count expected %v but received %v", 100, s.getCount())
	}
}

func TestCleaner(t *testing.T) {
	s := newStack()
	var nSlice []*node
	for i := 0; i < 100; i++ {
		nSlice = append(nSlice, s.Pop())
	}

	tn := getNow()
	for i := 0; i < 50; i++ {
		s.Push(nSlice[i], tn)
	}
	// Makes all the 50 pushed nodes invalid
	time.Sleep(time.Millisecond * 550)
	tn = getNow()
	for i := 50; i < 100; i++ {
		s.Push(nSlice[i], tn)
	}
	time.Sleep(time.Millisecond * 550)
	if s.getCount() != 50 {
		t.Fatalf("incorrect stack count expected %v but received %v", 50, s.getCount())
	}
	time.Sleep(time.Second)
	if s.getCount() != 0 {
		t.Fatalf("incorrect stack count expected %v but received %v", 0, s.getCount())
	}
}

// Display nodes for testing purposes
func (s *stack) Display() {
	for i := int32(0); i < s.getCount(); i++ {
		fmt.Printf("NODE IN STACK: %+v %p \n", s.nodes[i], s.nodes[i])
	}
	fmt.Println("TOTAL COUNT:", s.getCount())
}

//  158	   9,521,717 ns/op	 9600104 B/op	  100001 allocs/op
func BenchmarkWithoutStack(b *testing.B) {
	var n *node
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100000; j++ {
			n = new(node)
			n.value.Price = 1337
		}
	}
}

// 316	   3,485,211 ns/op	       1 B/op	       0 allocs/op
func BenchmarkWithStack(b *testing.B) {
	var n *node
	stack := newStack()
	b.ReportAllocs()
	b.ResetTimer()
	tn := getNow()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100000; j++ {
			n = stack.Pop()
			n.value.Price = 1337
			stack.Push(n, tn)
		}
	}
}

// getCount is a test helper function to derive the count that does not race.
func (s *stack) getCount() int32 {
	if !atomic.CompareAndSwapUint32(&s.sema, neutral, active) {
		return -1
	}
	defer atomic.StoreUint32(&s.sema, neutral)
	return s.count
}
