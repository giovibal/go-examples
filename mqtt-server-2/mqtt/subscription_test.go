package mqtt

import (
	"testing"
)

func BenchmarkSubscription_IsSubscribedEQ(b *testing.B) {
	s := NewSubscription("/test.it/jz/snapshot/snapshot/cis/1021", 0x00)
	for n := 0; n < b.N; n++ {
		s.IsSubscribed("/test.it/jz/snapshot/0000/cis/1021")
	}
}

func BenchmarkSubscription_IsSubscribedWC(b *testing.B) {
	s := NewSubscription("/+/jz/snapshot/+/cis/#", 0x00)
	for n := 0; n < b.N; n++ {
		s.IsSubscribed("/test.it/jz/snapshot/snapshot/cis/1021")
	}
}

func TestSubscription_IsSubscribed(t *testing.T) {
	s1 := NewSubscription("/+/jz/snapshot/+/cis/#", 0x00)
	b1, _ := s1.IsSubscribed("/test.it/jz/snapshot/snapshot/cis/1021")
	if !b1 {
		t.Fail()
	}

	s2 := NewSubscription("a", 0x00)
	b2, _ := s2.IsSubscribed("a")
	if !b2 {
		t.Fail()
	}
}
