package client

import (
	"testing"
	"time"
)

func TestKeyThrottled(t *testing.T) {
	k := BuildKey("hahaha", 1<<10, 1)
	start := time.Now()
	_, err := k.Use()
	if err != nil {
		t.Error("Failed at first attempt")
	}

	_, err = k.Use()
	after := time.Since(start)
	if err == nil {
		t.Errorf("Didn't get throttled after 2 tries in %v", after)
	}

	time.Sleep(time.Second * 1)

	_, err = k.Use()
	if err != nil {
		t.Errorf("Shouldn't be throttled after waiting for next second, %v", err)
	}

	k.Close()
}
