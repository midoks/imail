package smtpd

import (
	"fmt"
	"sync"
)

type FSMState string
type FSMEvent string
type FSMHandler func() FSMState

type FSM struct {
	mu       sync.Mutex
	state    FSMState
	handlers map[FSMState]map[FSMEvent]FSMHandler
}

func (f *FSM) getState() FSMState {
	return f.state
}

func (f *FSM) setState(newState FSMState) {
	f.state = newState
}

func (f *FSM) AddHandler(state FSMState, event FSMEvent, handler FSMHandler) *FSM {
	if _, ok := f.handlers[state]; !ok {
		f.handlers[state] = make(map[FSMEvent]FSMHandler)
	}
	if _, ok := f.handlers[state][event]; ok {
		fmt.Printf("[警告] 状态(%s)事件(%s)已定义过", state, event)
	}
	f.handlers[state][event] = handler
	return f
}

// 事件处理
func (f *FSM) Call(event FSMEvent) FSMState {
	f.mu.Lock()
	defer f.mu.Unlock()

	events := f.handlers[f.getState()]
	if events == nil {
		return f.getState()
	}

	if fn, ok := events[event]; ok {
		oldState := f.getState()
		f.setState(fn())
		newState := f.getState()
		fmt.Println("状态 [", oldState, "] 变成 [", newState, "]")
	}
	return f.getState()
}

func NewFSM(initState FSMState) *FSM {
	return &FSM{
		state:    initState,
		handlers: make(map[FSMState]map[FSMEvent]FSMHandler),
	}
}
