package main

import (
	"errors"
	"sync"
)

const (
	STATUS_SUCCEEDED = "SUCCEEDED"
	STATUS_RUNNING   = "RUNNING"
	STATUS_FAILED    = "FAILED"
)

type OrderProcess struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

var (
	ErrOrderProcessExists       = errors.New("order process exists")
	ErrOrderProcessDoesNotExist = errors.New("order process does not exist")
)

type OrderProcessRepository struct {
	OrderProcesses map[string]OrderProcess
	mu             sync.Mutex
}

func (o *OrderProcessRepository) Add(orderProcess OrderProcess) error {
	defer o.mu.Unlock()
	o.mu.Lock()

	if _, ok := o.OrderProcesses[orderProcess.OrderID]; ok {
		return ErrOrderProcessExists
	}

	o.OrderProcesses[orderProcess.OrderID] = orderProcess
	return nil
}

func (o *OrderProcessRepository) Update(orderProcess OrderProcess) error {
	_, err := o.GetByOrderID(orderProcess.OrderID)
	if err != nil {
		return err
	}

	o.OrderProcesses[orderProcess.OrderID] = orderProcess
	return nil
}

func (o *OrderProcessRepository) GetAll() ([]OrderProcess, error) {
	defer o.mu.Unlock()
	o.mu.Lock()

	orderProcesses := []OrderProcess{}
	for _, orderProcess := range o.OrderProcesses {
		orderProcesses = append(orderProcesses, orderProcess)
	}

	return orderProcesses, nil
}

func (o *OrderProcessRepository) GetByOrderID(orderID string) (OrderProcess, error) {
	defer o.mu.Unlock()
	o.mu.Lock()

	orderProcess, ok := o.OrderProcesses[orderID]
	if !ok {
		return OrderProcess{}, ErrOrderProcessDoesNotExist
	}

	return orderProcess, nil
}

func NewOrderProcessRepository() *OrderProcessRepository {
	return &OrderProcessRepository{
		OrderProcesses: make(map[string]OrderProcess),
	}
}
