// 实现 MemorySession 和 RedisSession
package session

type Session interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Del(key string) error
	Save() error
}