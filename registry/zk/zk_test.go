package zk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/internal/zk"
	"github.com/samaritan-proxy/sash/model"
)

func TestNewDiscoveryClient(t *testing.T) {
	t.Run("empty zk hosts", func(t *testing.T) {
		connCfg := &zk.ConnConfig{}
		_, err := NewDiscoveryClient(connCfg)
		assert.Error(t, err)
	})

	t.Run("empty base path", func(t *testing.T) {
		connCfg := &zk.ConnConfig{
			Hosts: []string{"127.0.0.1:2181"},
		}
		_, err := NewDiscoveryClient(connCfg)
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		connCfg := &zk.ConnConfig{
			Hosts:    []string{"127.0.0.1:2181"},
			BasePath: "/service",
		}
		c, err := NewDiscoveryClient(connCfg)
		assert.NoError(t, err)
		assert.NotNil(t, c)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			c.Run(ctx)
			close(done)
		}()
		time.AfterFunc(time.Millisecond*100, cancel)
		<-done
	})
}

type mockInstanceUnmarshaler struct{}

func (u *mockInstanceUnmarshaler) Unmarshal(data []byte) (inst *model.ServiceInstance, err error) {
	return
}

func TestDiscoveryClientOptions(t *testing.T) {
	t.Run("instance unmarshaler", func(t *testing.T) {
		instUnmarshaler := new(mockInstanceUnmarshaler)
		c, err := NewDiscoveryClientWithConn(
			nil,
			"/service",
			WithInstanceUnmarshaler(instUnmarshaler),
		)
		assert.NoError(t, err)
		assert.Equal(t, instUnmarshaler, c.instUnmarshaler)
	})
}

func TestDiscoveryClientListServiceNames(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	basePath := "/service"
	conn := zk.NewMockConn(ctrl)
	conn.EXPECT().Children(basePath).Return(
		[]string{"foo", "bar"},
		nil,
		nil,
	)

	c, err := NewDiscoveryClientWithConn(conn, "/service")
	assert.NoError(t, err)

	names, err := c.List()
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar"}, names)
}

func TestDiscoveryClientGetService(t *testing.T) {
	t.Run("non-existent service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := zk.NewMockConn(ctrl)
		conn.EXPECT().Children("/service/foo").Return(
			nil, nil, errors.New("node not exist"),
		)

		c, err := NewDiscoveryClientWithConn(conn, "/service")
		assert.NoError(t, err)

		_, err = c.Get("foo")
		assert.Error(t, err)
	})

	t.Run("non-existent instance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := zk.NewMockConn(ctrl)
		conn.EXPECT().Children("/service/foo").Return(
			[]string{"127.0.0.1_8888"}, nil, nil,
		)
		conn.EXPECT().Get("/service/foo/127.0.0.1_8888").Return(
			nil, nil, errors.New("node not exist"),
		)

		c, err := NewDiscoveryClientWithConn(conn, "/service")
		assert.NoError(t, err)

		_, err = c.Get("foo")
		assert.Error(t, err)
	})

	t.Run("invalid instance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := zk.NewMockConn(ctrl)
		conn.EXPECT().Children("/service/foo").Return(
			[]string{"127.0.0.1_8888"}, nil, nil,
		)
		conn.EXPECT().Get("/service/foo/127.0.0.1_8888").Return(
			[]byte(""), nil, nil,
		)

		c, err := NewDiscoveryClientWithConn(conn, "/service")
		assert.NoError(t, err)

		_, err = c.Get("foo")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := zk.NewMockConn(ctrl)
		conn.EXPECT().Children("/service/foo").Return(
			[]string{"127.0.0.1_8888", "127.0.0.1_8889"},
			nil, nil,
		)
		conn.EXPECT().Get("/service/foo/127.0.0.1_8888").Return(
			[]byte(`{"ip": "127.0.0.1", "port":8888}`),
			nil, nil,
		)
		conn.EXPECT().Get("/service/foo/127.0.0.1_8889").Return(
			[]byte(`{"ip": "127.0.0.1", "port":8889}`),
			nil, nil,
		)

		c, err := NewDiscoveryClientWithConn(conn, "/service")
		assert.NoError(t, err)

		svc, err := c.Get("foo")
		assert.NoError(t, err)
		assert.Len(t, svc.Instances, 2)
		assert.Contains(t, svc.Instances, "127.0.0.1:8888")
		assert.Contains(t, svc.Instances, "127.0.0.1:8889")
	})
}
