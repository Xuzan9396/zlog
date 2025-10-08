package zlog

import (
	"testing"
	"time"
)

func TestWatchErr(t *testing.T) {
	ch, err := WatchErr()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if allcancel != nil {
			allcancel()
		}
	}()

	go func() {
		for i := 0; i < 3; i++ {
			F().Errorf("watch err test: %d", i)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {
	case msg := <-ch:
		if msg == "" {
			t.Fatal("received empty message")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("watch err timeout")
	}
}

func TestWatchErrCallback(t *testing.T) {
	msgs := make(chan string, 1)

	if err := WatchErrCallback(func(msg string) {
		select {
		case msgs <- msg:
		default:
		}
	}); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if allcancel != nil {
			allcancel()
		}
	}()

	go func() {
		for i := 0; i < 3; i++ {
			F().Errorf("watch err callback test: %d", i)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {
	case msg := <-msgs:
		if msg == "" {
			t.Fatal("callback received empty message")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("watch err callback timeout")
	}
}
