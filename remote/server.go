/*
Copyright (C) 2018 KIM KeepInMind GmbH/srl

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
