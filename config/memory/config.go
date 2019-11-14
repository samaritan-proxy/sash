package memory

import (
	"github.com/samaritan-proxy/sash/config"
)

type configs struct {
	namespaces map[string]*namespace
}

func newConfigs() *configs {
	return &configs{namespaces: make(map[string]*namespace)}
}

func (c *configs) Exist(ns string) bool {
	_, ok := c.namespaces[ns]
	return ok
}

func (c *configs) Get(ns, typ, key string) ([]byte, error) {
	if !c.Exist(ns) {
		return nil, config.ErrNamespaceNotExist
	}
	return c.namespaces[ns].Get(typ, key)
}

func (c *configs) Set(ns, typ, key string, value []byte) {
	if !c.Exist(ns) {
		c.namespaces[ns] = newNamespace(ns)
	}
	c.namespaces[ns].Set(typ, key, value)
}

func (c *configs) Del(ns, typ, key string) error {
	if !c.Exist(ns) {
		return config.ErrNamespaceNotExist
	}
	namespace := c.namespaces[ns]
	if err := namespace.Del(typ, key); err != nil {
		return err
	}
	if namespace.IsEmpty() {
		delete(c.namespaces, ns)
	}
	return nil
}

func (c *configs) Keys(ns, typ string) ([]string, error) {
	if !c.Exist(ns) {
		return nil, config.ErrNamespaceNotExist
	}
	return c.namespaces[ns].Keys(typ)
}

type namespace struct {
	name  string
	types map[string]*typ
}

func newNamespace(name string) *namespace {
	return &namespace{
		name:  name,
		types: make(map[string]*typ),
	}
}

func (n *namespace) Exist(typ string) bool {
	_, ok := n.types[typ]
	return ok
}

func (n *namespace) Get(typ, key string) ([]byte, error) {
	if !n.Exist(typ) {
		return nil, config.ErrTypeNotExist
	}
	return n.types[typ].Get(key)
}

func (n *namespace) Set(typ, key string, value []byte) {
	if !n.Exist(typ) {
		n.types[typ] = newType(typ)
	}
	n.types[typ].Set(key, value)
}

func (n *namespace) Del(typ, key string) error {
	if !n.Exist(typ) {
		return config.ErrTypeNotExist
	}
	typs := n.types[typ]
	if err := typs.Del(key); err != nil {
		return err
	}
	if typs.IsEmpty() {
		delete(n.types, typ)
	}
	return nil
}

func (n *namespace) Keys(typ string) ([]string, error) {
	if !n.Exist(typ) {
		return nil, config.ErrTypeNotExist
	}
	return n.types[typ].Keys(), nil
}

func (n *namespace) IsEmpty() bool {
	return len(n.types) == 0
}

type typ struct {
	name    string
	configs map[string][]byte
}

func newType(name string) *typ {
	return &typ{
		name:    name,
		configs: make(map[string][]byte),
	}
}

func (t *typ) Exist(key string) bool {
	_, ok := t.configs[key]
	return ok
}

func (t *typ) Get(key string) ([]byte, error) {
	if !t.Exist(key) {
		return nil, config.ErrKeyNotExist
	}
	return t.configs[key], nil
}

func (t *typ) Set(key string, value []byte) {
	t.configs[key] = value
}

func (t *typ) Del(key string) error {
	if !t.Exist(key) {
		return config.ErrKeyNotExist
	}
	delete(t.configs, key)
	return nil
}

func (t *typ) Keys() []string {
	keys := make([]string, 0, len(t.configs))
	for k := range t.configs {
		keys = append(keys, k)
	}
	return keys
}

func (t *typ) IsEmpty() bool {
	return len(t.configs) == 0
}
