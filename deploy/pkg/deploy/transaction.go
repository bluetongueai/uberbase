package deploy

import (
	"fmt"
	"time"
)

type TransactionStatus string

const (
	TxStarted   TransactionStatus = "started"
	TxCompleted TransactionStatus = "completed"
	TxFailed    TransactionStatus = "failed"
	TxRollback  TransactionStatus = "rollback"
)

type TransactionLog struct {
	ID          string            `yaml:"id"`
	ServiceName string            `yaml:"service_name"`
	Action      string            `yaml:"action"`
	Status      TransactionStatus `yaml:"status"`
	Timestamp   time.Time         `yaml:"timestamp"`
	Error       string            `yaml:"error,omitempty"`
	Version     string            `yaml:"version"`
	Metadata    map[string]string `yaml:"metadata,omitempty"`
}

type TransactionManager struct {
	stateManager *StateManager
}

func NewTransactionManager(stateManager *StateManager) *TransactionManager {
	return &TransactionManager{
		stateManager: stateManager,
	}
}

func (t *TransactionManager) LogTransaction(log TransactionLog) error {
	state, err := t.stateManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.Services[log.ServiceName] == nil {
		state.Services[log.ServiceName] = &ServiceState{}
	}

	currentState := state.Services[log.ServiceName]
	if currentState.Transactions == nil {
		currentState.Transactions = []TransactionLog{}
	}

	// Generate unique ID if not provided
	if log.ID == "" {
		log.ID = fmt.Sprintf("tx-%d", time.Now().UnixNano())
	}

	currentState.Transactions = append(currentState.Transactions, log)
	return t.stateManager.Save(state)
}

func (t *TransactionManager) GetLastTransaction(serviceName string) (*TransactionLog, error) {
	state, err := t.stateManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	if state.Services[serviceName] == nil ||
		len(state.Services[serviceName].Transactions) == 0 {
		return nil, nil
	}

	transactions := state.Services[serviceName].Transactions
	return &transactions[len(transactions)-1], nil
}

func (t *TransactionManager) GetFailedTransactions() (map[string]*TransactionLog, error) {
	state, err := t.stateManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	failed := make(map[string]*TransactionLog)
	for serviceName, serviceState := range state.Services {
		if len(serviceState.Transactions) > 0 {
			lastTx := serviceState.Transactions[len(serviceState.Transactions)-1]
			if lastTx.Status == TxFailed {
				failed[serviceName] = &lastTx
			}
		}
	}

	return failed, nil
}
