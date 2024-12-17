package question

import (
	"sync"
)

// ConcurrentMap  使用互斥锁实现 并发安全的 map
type ConcurrentMap struct {
	// 并发量
	data map[string]interface{}
	lock sync.RWMutex
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		data: make(map[string]interface{}),
		lock: sync.RWMutex{},
	}
}
func (m *ConcurrentMap) Get(key string) (interface{}, bool) {
	// 读锁
	m.lock.RLock()
	defer m.lock.RUnlock()
	value, ok := m.data[key]
	return value, ok
}

func (m *ConcurrentMap) Set(key string, value interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[key] = value
}

func (m *ConcurrentMap) Delete(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.data, key)
}

func (m *ConcurrentMap) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.data)
}

// ConcurrentMapByChannel  使用 channel 实现的并发安全的 map
type ConcurrentMapByChannel struct {
	commands chan command
	closed   chan struct{}
}

type command interface {
	execute(map[string]interface{})
}

type setCommand struct {
	key   string
	value interface{}
}

func (command *setCommand) execute(data map[string]interface{}) {
	data[command.key] = command.value
}

type getCommand struct {
	key      string
	response chan<- interface{}
}

func (command *getCommand) execute(data map[string]interface{}) {
	command.response <- data[command.key]
}

type deleteCommand struct {
	key string
}

func (command *deleteCommand) execute(data map[string]interface{}) {
	delete(data, command.key)
}

type lenCommand struct {
	response chan<- int
}

func (command *lenCommand) execute(data map[string]interface{}) {
	command.response <- len(data)
}

// NewConcurrentMapByChannel 创建一个并发安全的 map。
func NewConcurrentMapByChannel() *ConcurrentMapByChannel {
	cm := &ConcurrentMapByChannel{
		commands: make(chan command, 100),
		closed:   make(chan struct{}),
	}
	go cm.run()
	return cm
}

func (cm *ConcurrentMapByChannel) run() {
	data := make(map[string]interface{})
	for {
		select {
		case <-cm.closed:
			close(cm.commands)
			return
		case cmd := <-cm.commands:
			cmd.execute(data)
		}
	}
}

func (cm *ConcurrentMapByChannel) Set(key string, value interface{}) {
	cm.commands <- &setCommand{key, value}
}

func (cm *ConcurrentMapByChannel) Get(key string) interface{} {
	response := make(chan interface{})
	cm.commands <- &getCommand{key, response}
	return <-response
}

func (cm *ConcurrentMapByChannel) Delete(key string) {
	cm.commands <- &deleteCommand{key}
}

func (cm *ConcurrentMapByChannel) Len() int {
	response := make(chan int)
	cm.commands <- &lenCommand{response}
	return <-response
}

func (cm *ConcurrentMapByChannel) Close() {
	close(cm.closed)
}
