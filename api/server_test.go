package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/config/memory"
	"github.com/samaritan-proxy/sash/registry"
)

func newTestServer(t *testing.T, mockCtl *gomock.Controller, opts ...ServerOption) *Server {
	reg := registry.NewMockCache(mockCtl)
	ctl := config.NewController(memory.NewMemStore(), config.Interval(time.Millisecond))
	assert.NoError(t, ctl.Start())
	return New(reg, ctl, opts...)
}

func testHandler(req *http.Request, server *Server) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	server.hs.Handler.ServeHTTP(resp, req)
	return resp
}

func assertDoNotTimeout(t *testing.T, fn func(), d time.Duration) {
	done := make(chan struct{})
	timer := time.NewTimer(d)
	defer timer.Stop()

	go func() {
		fn()
		close(done)
	}()

	select {
	case <-done:
	case <-timer.C:
		t.Fatal("timeout")
	}
}

func TestAddr(t *testing.T) {
	options := new(serverOptions)
	Addr(":80")(options)
	assert.Equal(t, ":80", options.Addr)
}

func TestReadTimeout(t *testing.T) {
	options := new(serverOptions)
	ReadTimeout(time.Second * 11)(options)
	assert.Equal(t, time.Second*11, options.ReadTimeout)
}

func TestReadHeaderTimeout(t *testing.T) {
	options := new(serverOptions)
	ReadHeaderTimeout(time.Second * 13)(options)
	assert.Equal(t, time.Second*13, options.ReadHeaderTimeout)
}

func TestWriteTimeout(t *testing.T) {
	options := new(serverOptions)
	WriteTimeout(time.Second * 17)(options)
	assert.Equal(t, time.Second*17, options.WriteTimeout)
}

func TestIdleTimeout(t *testing.T) {
	options := new(serverOptions)
	IdleTimeout(time.Second * 19)(options)
	assert.Equal(t, time.Second*19, options.IdleTimeout)
}

func TestServer(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	s := newTestServer(t, ctl, Addr("127.0.0.1:18882"))

	assert.NoError(t, s.Start())

	resp, err := http.Get("http://127.0.0.1:18882/ping")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assertDoNotTimeout(t, s.Stop, time.Second)
}
