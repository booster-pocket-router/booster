package remote_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/booster-proj/booster/remote"
)

func TestListenAndServe(t *testing.T) {
	srv := remote.New(http.DefaultServeMux)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan error)
	go func() {
		c <- srv.ListenAndServe(ctx, 0)
	}()

	cancel()
	select {
	case <-time.After(100 * time.Millisecond):
		t.Fatal("shutdown timeout")
	case <-c:
	}
}
