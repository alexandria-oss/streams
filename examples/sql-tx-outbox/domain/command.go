package domain

import (
	"context"
	"errors"
	"sync"

	"github.com/modern-go/reflect2"
)

var (
	ErrCommandAlreadyExists = errors.New("command_bus: command already exists")
	ErrCommandNotFound      = errors.New("command_bus: command not found")
)

type CommandHandlerFunc func(ctx context.Context, cmd any) error

type CommandBus interface {
	Exec(ctx context.Context, cmd any) error
}

type SyncCommandBus interface {
	CommandBus
	Register(cmd any, handlerFunc CommandHandlerFunc) error
}

type AsyncCommandBus interface {
	CommandBus
	Register(streamName string, cmd any, handlerFunc CommandHandlerFunc) error
}

type MemoryCommandBus struct {
	mu  sync.RWMutex
	reg map[string]CommandHandlerFunc
}

var _ SyncCommandBus = &MemoryCommandBus{}

func NewMemoryCommandBus() *MemoryCommandBus {
	return &MemoryCommandBus{
		mu:  sync.RWMutex{},
		reg: map[string]CommandHandlerFunc{},
	}
}

func (m *MemoryCommandBus) Register(cmd any, handlerFunc CommandHandlerFunc) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmdType := reflect2.TypeOf(cmd).String()
	if _, ok := m.reg[cmdType]; ok {
		return ErrCommandAlreadyExists
	}

	m.reg[cmdType] = handlerFunc
	return nil
}

func (m *MemoryCommandBus) Exec(ctx context.Context, cmd any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmdType := reflect2.TypeOf(cmd)
	cmdHandler, ok := m.reg[cmdType.String()]
	if !ok {
		return ErrCommandNotFound
	}

	return cmdHandler(ctx, cmd)
}
