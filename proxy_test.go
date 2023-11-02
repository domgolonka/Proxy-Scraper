package proxy

import (
	"testing"
	"time"
)

func ProxyTest(t *testing.T) {
	t.Run("ProxyTest", func(t *testing.T) {
		gen := New(1, 5*time.Second, 5, nil)
		time.Sleep(10 * time.Second)
		if gen.Count() == 0 {
			t.Errorf("Expected %v got %v", 0, gen.Count())
		}
	})
}
