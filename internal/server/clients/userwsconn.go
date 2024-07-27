package clients

import (
	"sync"

	"github.com/gorilla/websocket"
)

type UserConnMap struct {
	mu    *sync.RWMutex
	value map[int64][]*websocket.Conn
}

func NewWSConnMap() *UserConnMap {
	return &UserConnMap{
		mu:    &sync.RWMutex{},
		value: make(map[int64][]*websocket.Conn),
	}
}

func (m *UserConnMap) Put(userId int64, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.value[userId] = append(m.value[userId], conn)
}

func (m *UserConnMap) UserCons(userId int64) []*websocket.Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.value[userId]
}
