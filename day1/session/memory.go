package session

import (
	"errors"
	"sync"
)

// 对象
// MemorySession 设计
// 定义 MemorySession 对象（字段：sessionId, 存 kv 的 map, 读写锁）
// 构造函数，为了获取对象
// Set()
// Get()
// Del()
// Save()
type MemorySession struct {
	sessionId string
	// 存kv
	data   map[string]interface{}
	rwlock sync.RWMutex
}

// 构造函数
func NewMemorySession(id string) *MemorySession {
	s := &MemorySession{
		sessionId: id,
		data:      make(map[string]interface{}, 16),
	}
	return s
}

func (m *MemorySession) Set(key string, value interface{}) (err error) {
	// 写入加锁
	m.rwlock.Lock()
	defer m.rwlock.Unlock()
	// 设置值
	m.data[key] = value
	return
}

func (m *MemorySession) Get(key string) (value interface{}, err error) {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()
	value, ok := m.data[key]
	if !ok {
		err = errors.New("key not exists in session")
		return
	}
	return
}

func (m *MemorySession) Del(key string) (err error) {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()
	delete(m.data, key)
	return
}

func (m *MemorySession) Save() (err error) {
	return
}
