package memstorage

import (
	"context"	
	"fmt"
	"sync"

	"github.com/rebus2015/gophermart/cmd/internal/logger"
	"github.com/rebus2015/gophermart/cmd/internal/model"
)

type OrdersMux struct {
	values map[int64]model.Order
	mux    sync.RWMutex
}

type MemStorage struct {
	orders OrdersMux
	db     dbStorage
	cfg    config
	lg     *logger.Logger
}

type config interface {
	GetAccruralAddr() string
}

type dbStorage interface {
	AccruralUpdate(order *model.Order) error
	OrdersAcc() (*[]model.Order, error)
}

func NewStorage(ctg context.Context, db dbStorage, cfg config, lg *logger.Logger) *MemStorage {
	m := &MemStorage{
		orders: OrdersMux{
			values: map[int64]model.Order{},
			mux:    sync.RWMutex{},
		},
		db:  db,
		cfg: cfg,
		lg:  lg,
	}
	return m
}

func (m *MemStorage) Restore() error {
	orders, err := m.db.OrdersAcc()
	if err != nil {
		m.lg.Fatal().Err(err).Msgf("Failed to restore orders for MemStorage err:%v", err)
		return err
	}
	if len(*orders) == 0 {
		return nil
	}
	m.orders.mux.Lock()
	defer m.orders.mux.Unlock()
	for _, o := range *orders {
		m.orders.values[*o.Num] = o
	}
	m.lg.Debug().Msgf("MemStorage restored successfully with [%v] orders", len(m.orders.values))
	return nil
}

func (m *MemStorage) Add(order *model.Order) {
	m.orders.mux.Lock()
	defer m.orders.mux.Unlock()
	m.orders.values[*order.Num] = *order
	m.lg.Debug().Msgf("MemStorage adder order number %v", order.Num)
}

func (m *MemStorage) Update(order *model.Order) error{
	m.orders.mux.Lock()
	defer m.orders.mux.Unlock()
	if _, ok := m.orders.values[*order.Num]; !ok {
		m.lg.Error().Msgf("Order number [%v] not found on memStorare", order.Num)
		return fmt.Errorf("Order number [%v] not found on memStorare", order.Num)
	}

	err := m.db.AccruralUpdate(order)
	if err != nil {
		m.lg.Err(err).Msgf("[Memstorage.Update] Error. Failed to update accrual for order [%v]: %v", order.Num, err)
		return fmt.Errorf("[Memstorage.Update] Error. Failed to update accrual for order [%v]: %v", order.Num, err)
	}
	delete(m.orders.values, *order.Num)
	m.lg.Debug().Msgf("[MemStorage] order number %v UPDATED", order.Num)
	return nil
}

func (m *MemStorage) List() []*model.Order{
	m.orders.mux.RLock()
	defer m.orders.mux.RUnlock()
	oslice := make([]*model.Order, 0, len(m.orders.values))
    for _, tx := range m.orders.values {
        oslice = append(oslice, &tx)
    }
	return oslice
}