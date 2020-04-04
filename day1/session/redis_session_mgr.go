package session

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

//RedisSessionMgr 设计:
//定义 RedisSessionMgr 对象（字段：redis 地址, redis 密码, 连接池, 读写锁, 大 map）
//构造函数
//Init()
//CreateSession()
//GetSession()

type RedisSessionMgr struct {
	// redis 地址
	addr string
	// 密码
	passwd string
	// 连接池
	pool *redis.Pool
	// 锁
	rwLock sync.RWMutex
	// 大map
	sessionMap map[string]Session
}

// 构造
func NewRedisSessionMgr(addr, passwd string, pool *redis.Pool, rwLock sync.RWMutex, sessionMap map[string]Session) *RedisSessionMgr {
	return &RedisSessionMgr{
		addr:       "",
		passwd:     "",
		pool:       nil,
		rwLock:     sync.RWMutex{},
		sessionMap: make(map[string]Session, 16),
	}
}

func (r *RedisSessionMgr) Init(addr string, options ...string) (err error) {
	// 若有其他参数
	if len(options) > 0 {
		r.passwd = options[0]
	}
	// 创建连接池
	r.pool = myPool(addr, r.passwd)
	return nil
}

func myPool(addr, password string) *redis.Pool {
	return &redis.Pool{
		Dial: func() (conn redis.Conn, e error) {
			conn, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			// 若有密码，判断
			if _, err := conn.Do("AUTH", password); err != nil {
				conn.Close()
				return nil, err
			}
			return
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		MaxIdle:     64,
		MaxActive:   1000,
		IdleTimeout: time.Second,
		Wait:        false,
	}
}

func (r *RedisSessionMgr) CreateSession() (session Session, err error) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	// 用uuid作为sessionId
	u := uuid.NewV4()
	sessionId := u.String()
	// 创建session
	session = NewRedisSession(sessionId, r.pool)
	// 加入到大map
	r.sessionMap[sessionId] = session
	return
}

func (r *RedisSessionMgr) Get(sessionId string) (session Session, err error) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	session, ok := r.sessionMap[sessionId]
	if !ok {
		err = errors.New("session not exists")
		return
	}
	return
}
