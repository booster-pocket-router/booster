package remote

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Remote struct {
	*http.Server
}

func New(h http.Handler) *Remote {
	return &Remote{
		&http.Server{
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      h,
		},
	}
}

func (r *Remote) ListenAndServe(ctx context.Context, port int) error {
	c := make(chan error)
	go func() {
		r.Server.Addr = fmt.Sprintf(":%d", port)
		c <- r.Server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		r.Shutdown(ctx)
		return <-c
	case err := <-c:
		return err
	}
}
