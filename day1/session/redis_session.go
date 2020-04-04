package session

import (
	"encoding/json"
	"errors"
	"github.com/garyburd/redigo/redis"
	"sync"
)

//RedisSession 设计:
//定义 RedisSession 对象（字段: sessionId, 存kv的map, 读写锁, redis连接池, 记录内存中 map 是否被修改的标记）
//构造函数
//Set(): 将 session 存到内存中的 map
//Get(): 取数据，实现延迟加载
//Del()
//Save(): 将 session 存到 redis

type RedisSession struct {
	sessionId string
	pool      *redis.Pool
	// 设置 session, 可以先放在内存的 map 中
	// 批量的导入 redis, 提升性能
	sessionMap map[string]interface{}
	// 读写锁
	rwLock sync.RWMutex
	// 记录内存中 map 是否被操作
	flag int
}

const (
	// 内存数据没变化
	SessionFlagNone = iota
	// 内存数据由变化
	SessionFlagModify
)

// 构造函数
func NewRedisSession(id string, pool *redis.Pool) *RedisSession {
	return &RedisSession{
		sessionId:  id,
		pool:       pool,
		sessionMap: make(map[string]interface{}, 16),
		rwLock:     sync.RWMutex{},
		flag:       SessionFlagNone,
	}
}

func (r *RedisSession) Set(key string, value interface{}) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	// 设置值
	r.sessionMap[key] = value
	// 标记记录
	r.flag = SessionFlagModify
	return nil
}

func (r *RedisSession) Get(key string) (interface{}, error) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	// 先判断内存
	session, ok := r.sessionMap[key]
	if !ok {
		err := errors.New("key not exists")
		return nil, err
	}
	return session, nil
}

// 从redis里再次加载
func (r *RedisSession) loadFromRedis() error {
	conn := r.pool.Get()
	reply, err := conn.Do("GET", r.sessionId)
	if err != nil {
		return err
	}
	data, err := redis.String(reply, err)
	if err != nil {
		return err
	}
	// 取到的东西，反序列化到内存的map
	err = json.Unmarshal([]byte(data), &r.sessionMap)
	if err != nil {
		return err
	}
	return err
}

func (r *RedisSession) Del(key string) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	r.flag = SessionFlagModify
	delete(r.sessionMap, key)
	return nil
}

func (r *RedisSession) Save() error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	// 若数据没变，不需要存
	if r.flag != SessionFlagModify {
		return nil
	}
	// 内存中的sessionMap进行序列化
	data, err := json.Marshal(r.sessionMap)
	if err != nil {
		return err
	}
	// 获取redis连接
	conn := r.pool.Get()
	// 保存kv
	_, err = conn.Do("SET", r.sessionId, string(data))
	if err != nil {
		return err
	}
	return err
}
