package memstorage

import (
	"context"
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

func (m *MemStorage) Update(number int64) {
	m.orders.mux.Lock()
	defer m.orders.mux.Unlock()
	if _, ok := m.orders.values[number]; !ok {
		m.lg.Error().Msgf("Order number [%v] not found on memStorare", number)
		return
	}
	order := m.orders.values[number]
	err := m.db.AccruralUpdate(&order)
	if err != nil {
		m.lg.Err(err).Msgf("[Memstorage.Update] Error. Failed to update accrual for order [%v]: %v", order.Num, err)
		return
	}
	delete(m.orders.values, number)
	m.lg.Debug().Msgf("[MemStorage] order number %v UPDATED", order.Num)
}
