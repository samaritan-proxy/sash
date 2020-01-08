package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/config"
	cfgmem "github.com/samaritan-proxy/sash/config/memory"
	"github.com/samaritan-proxy/sash/registry"
	regmem "github.com/samaritan-proxy/sash/registry/memory"
)

func newTestServer(t *testing.T, opts ...ServerOption) *Server {
	reg := registry.NewCache(regmem.NewRegistry())
	ctl := config.NewController(cfgmem.NewMemStore(), config.Interval(time.Millisecond))
	assert.NoError(t, ctl.Start())
	return New("127.0.0.1:18882", reg, ctl, opts...)
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
	s := newTestServer(t)

	go func() {
		assert.NoError(t, s.Start())
	}()

	resp, err := http.Get("http://127.0.0.1:18882/api/ping")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assertDoNotTimeout(t, s.Stop, time.Second)
}
