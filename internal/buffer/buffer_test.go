// internal/buffer/buffer_test.go
package buffer_test

import (
	"testing"

	"xteve/internal/buffer"
)

func TestTunerLimit(t *testing.T) {
	b := buffer.New(buffer.Config{TunerCount: 2})

	// Acquire up to limit
	id1, err := b.Acquire("http://stream1")
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	id2, err := b.Acquire("http://stream2")
	if err != nil {
		t.Fatalf("second Acquire: %v", err)
	}

	// Third should fail
	_, err = b.Acquire("http://stream3")
	if err == nil {
		t.Error("expected error at tuner limit, got nil")
	}
	if err != buffer.ErrTunerLimitReached {
		t.Errorf("error = %v, want ErrTunerLimitReached", err)
	}

	// Release one, then acquire should succeed
	b.Release(id1)
	_, err = b.Acquire("http://stream3")
	if err != nil {
		t.Fatalf("Acquire after release: %v", err)
	}
	_ = id2
}

func TestActiveCount(t *testing.T) {
	b := buffer.New(buffer.Config{TunerCount: 5})

	if b.ActiveCount() != 0 {
		t.Errorf("initial active = %d, want 0", b.ActiveCount())
	}

	id, _ := b.Acquire("http://stream1")
	if b.ActiveCount() != 1 {
		t.Errorf("active after acquire = %d, want 1", b.ActiveCount())
	}

	b.Release(id)
	if b.ActiveCount() != 0 {
		t.Errorf("active after release = %d, want 0", b.ActiveCount())
	}
}
