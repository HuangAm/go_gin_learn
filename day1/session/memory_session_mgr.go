package session

import (
	"errors"
	"github.com/satori/go.uuid"
	"sync"
)

//MemorySessionMgr 设计:
//定义 MemorySessionMgr 对象（字段: 存放所有 session 的map, 读写锁）
//构造函数
//Init()
//CreateSession()
//GetSession()
type MemorySessionMgr struct {
	sessionMap map[string]Session
	rwLock     sync.RWMutex
}

func NewMemorySessionMgr() *MemorySessionMgr {
	sr := &MemorySessionMgr{
		sessionMap: make(map[string]Session, 1024),
	}
	return sr
}

func (s *MemorySessionMgr) Init(addr string, options ...string) (err error) {
	return
}

func (s *MemorySessionMgr) CreateSession() (session Session, err error) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	u := uuid.NewV4()
	// UUID 转 string
	sessionId := u.String()
	// 创建 session
	session = NewMemorySession(sessionId)
	// 加入到大 map
	s.sessionMap[sessionId] = session
	return
}

func (s *MemorySessionMgr) Get(sessionId string) (session Session, err error) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	session, ok := s.sessionMap[sessionId]
	if !ok{
		err = errors.New("session not exists")
		return
	}
	return
}
