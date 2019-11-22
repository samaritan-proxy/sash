package zk

import (
	"testing"

	"github.com/golang/mock/gomock"
	zkpkg "github.com/mesosphere/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/internal/zk"
)

func TestNewWithConn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty base path", func(t *testing.T) {
		conn := zk.NewMockConn(ctrl)
		_, err := NewWithConn(conn, "")
		assert.Error(t, err)
	})

	t.Run("correct", func(t *testing.T) {
		conn := zk.NewMockConn(ctrl)
		_, err := NewWithConn(conn, "/configs")
		assert.NoError(t, err)
	})
}

func TestStore_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := zk.NewMockConn(ctrl)
	conn.EXPECT().Get(gomock.Not(gomock.Eq("/configs/ns/type/key"))).Return(nil, nil, zkpkg.ErrNoNode)
	conn.EXPECT().Get("/configs/ns/type/key").Return([]byte("value"), nil, nil)

	t.Run("no node error", func(t *testing.T) {
		s, err := NewWithConn(conn, "/configs")
		assert.NoError(t, err)
		_, err = s.Get("ns", "type", "key1")
		assert.Equal(t, config.ErrNotExist, err)
	})

	t.Run("correct", func(t *testing.T) {
		s, err := NewWithConn(conn, "/configs")
		assert.NoError(t, err)
		value, err := s.Get("ns", "type", "key")
		assert.NoError(t, err)
		assert.Equal(t, []byte("value"), value)
	})
}

func TestStore_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := zk.NewMockConn(ctrl)
	conn.EXPECT().CreateRecursively("/configs/ns/type/key", []byte("value")).Return(nil)
	s, err := NewWithConn(conn, "/configs")
	assert.NoError(t, err)
	assert.NoError(t, s.Set("ns", "type", "key", []byte("value")))
}

func TestStore_Del(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := zk.NewMockConn(ctrl)
	conn.EXPECT().DeleteWithChildren(gomock.Not(gomock.Eq("/configs/ns/type/key"))).Return(zkpkg.ErrNoNode)
	conn.EXPECT().DeleteWithChildren("/configs/ns/type/key").Return(nil)

	t.Run("no node error", func(t *testing.T) {
		s, err := NewWithConn(conn, "/configs")
		assert.NoError(t, err)
		assert.Equal(t, config.ErrNotExist, s.Del("ns", "type", "key1"))
	})

	t.Run("correct", func(t *testing.T) {
		s, err := NewWithConn(conn, "/configs")
		assert.NoError(t, err)
		assert.NoError(t, s.Del("ns", "type", "key"))
	})
}

func TestStore_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := zk.NewMockConn(ctrl)
	conn.EXPECT().Exists(gomock.Not(gomock.Eq("/configs/ns/type/key"))).Return(false, nil, nil)
	conn.EXPECT().Exists("/configs/ns/type/key").Return(true, nil, nil)

	t.Run("false", func(t *testing.T) {
		s, err := NewWithConn(conn, "/configs")
		assert.NoError(t, err)
		assert.False(t, s.Exist("ns", "type", "key1"))
	})

	t.Run("true", func(t *testing.T) {
		s, err := NewWithConn(conn, "/configs")
		assert.NoError(t, err)
		assert.True(t, s.Exist("ns", "type", "key"))
	})
}

func TestStore_GetKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := zk.NewMockConn(ctrl)
	conn.EXPECT().Children(gomock.Not(gomock.Eq("/configs/ns/type"))).Return(nil, nil, zkpkg.ErrNoNode)
	conn.EXPECT().Children("/configs/ns/type").Return([]string{"key1", "key2"}, nil, nil)

	t.Run("no node error", func(t *testing.T) {
		s, err := NewWithConn(conn, "/configs")
		assert.NoError(t, err)
		_, err = s.GetKeys("ns", "type1")
		assert.Equal(t, config.ErrNotExist, err)
	})

	t.Run("correct", func(t *testing.T) {
		s, err := NewWithConn(conn, "/configs")
		assert.NoError(t, err)
		keys, err := s.GetKeys("ns", "type")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
	})
}
