package executor

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type Executor struct {
	adapters map[string]Adapter
	mu       sync.RWMutex
	logger   *zap.Logger
}

func NewExecutor(logger *zap.Logger) *Executor {
	return &Executor{
		adapters: make(map[string]Adapter),
		logger:   logger,
	}
}

func (e *Executor) RegisterAdapter(adapter Adapter) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.adapters[adapter.Name()] = adapter
	e.logger.Info("registered adapter", zap.String("name", adapter.Name()))
}

func (e *Executor) GetAdapter(name string) (Adapter, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	adapter, ok := e.adapters[name]
	if !ok {
		return nil, fmt.Errorf("adapter %q not registered", name)
	}
	return adapter, nil
}

func (e *Executor) Deploy(ctx context.Context, adapterName string, req DeployRequest) (*DeployResponse, error) {
	adapter, err := e.GetAdapter(adapterName)
	if err != nil {
		return nil, err
	}
	return adapter.Deploy(ctx, req)
}

func (e *Executor) Stop(ctx context.Context, adapterName string, req StopRequest) error {
	adapter, err := e.GetAdapter(adapterName)
	if err != nil {
		return err
	}
	return adapter.Stop(ctx, req)
}
