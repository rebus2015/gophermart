package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rebus2015/gophermart/cmd/internal/client/agent"
	"github.com/rebus2015/gophermart/cmd/internal/logger"
	"github.com/rebus2015/gophermart/cmd/internal/model"
)

type AccrualClient struct {
	storage dbStorage
	cfg     config
	lg      *logger.Logger
	ctx     context.Context
	client  *http.Client
}

type config interface {
	GetAccruralAddr() string
	GetSyncInterval() time.Duration
	GetRateLimit() int
}

type dbStorage interface {
	AccruralUpdate(order *model.Accrual) error
	OrdersAcc() ([]model.Order, error)
}

func NewClient(c context.Context, s dbStorage, conf config, logger *logger.Logger) *AccrualClient {
	return &AccrualClient{
		storage: s,
		cfg:     conf,
		lg:      logger,
		ctx:     c,
		client: &http.Client{
			Transport: &http.Transport{IdleConnTimeout: 5 * time.Second},
		},
	}
}

func (ac *AccrualClient) Run() {
	errCh := make(chan error) // создаём канал, из которого будем ждать ошибку
	go ac.sndWorker(errCh)
}

func (ac *AccrualClient) sndWorker(errCh chan<- error) {
	ticker := time.NewTicker(ac.cfg.GetSyncInterval())
	defer close(errCh)
	for {
		select {
		case <-ticker.C:
			err := ac.updateSendMultiple()
			if err != nil {
				errCh <- fmt.Errorf("error update orders: %w", err)
			}
		case <-ac.ctx.Done():
			ac.lg.Info().Msgf("request worker stopped")
			return
		}
	}
}

func (ac *AccrualClient) updateSendMultiple() error {
	orders, err := ac.storage.OrdersAcc()
	if err != nil {
		ac.lg.Err(err).Msg("Client failed to Update Orders List")
		return err
	}
	jobs := []agent.Job{}
	length := len(orders)
	for i := 0; i < length; i++ {
		jobs = append(jobs, agent.Job{
			Descriptor: i,
			ExecFn:     ac.sendreq,
			Args: agent.Args{
				Order: &orders[i],
			},
		})
	}
	wp := agent.New(ac.cfg.GetRateLimit(), ac.lg)

	go wp.GenerateFrom(jobs)
	go wp.Run(ac.ctx)

	for {
		select {
		case r, ok := <-wp.ErrCh():
			if !ok {
				continue
			}
			if r.Err != nil {
				//time.Sleep(time.Minute * 1)
				ac.lg.Printf("unexpected error: %v from worker on Job %v", r.Err, r.Descriptor)
			}
			ac.lg.Printf("worker processed Job %v", r.Descriptor)

		case <-wp.Done:
			ac.lg.Printf("worker FINISHED")
			return nil
		}
	}
}

func (ac *AccrualClient) sendreq(ctx context.Context, args agent.Args) error {
	queryurl := ac.cfg.GetAccruralAddr() + "/api/orders/" + strconv.FormatInt(*args.Order.Num, 10)
	ac.lg.Info().Msgf("Create Request Url: %s", queryurl)

	r, err := http.NewRequestWithContext(ac.ctx, http.MethodGet, queryurl, nil)
	if err != nil {
		ac.lg.Err(err).Msgf("Create Request failed! with error: %v\n", err)
		return err
	}

	response, err := ac.client.Do(r)
	if err != nil {
		ac.lg.Err(err).Msgf("Send request error: %v", err)
		return err
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		{
			ac.lg.Info().Msgf("[AccrualService] responce status [%v] for order [%v]", response.StatusCode, *args.Order.Num)
			order := &model.Accrual{}
			decoder := json.NewDecoder(response.Body)
			if err := decoder.Decode(order); err != nil {
				ac.lg.Printf("Read response body error: %v", err)
				return err
			}
			result, _ := json.Marshal(order)
			ac.lg.Info().Msgf("[Client] updating Accrual for order: %v", string(result))
			err = ac.storage.AccruralUpdate(order)
			if err != nil {
				ac.lg.Err(err).Msgf("Failed to update order info [%v]", order)
			}
		}
	case http.StatusNoContent:
		{
			ac.lg.Info().Msgf("[AccrualService] responce status [%v] for order [%v]", response.StatusCode, *args.Order.Num)
			_, err := io.ReadAll(response.Body)
			if err != nil {
				ac.lg.Printf("Read response body error: %v", err)
				return err
			}
			return fmt.Errorf("[AccrualService] responce error, status [%v] for order [%v]", response.StatusCode, *args.Order.Num)
		}
	case http.StatusTooManyRequests:
		{
			ac.lg.Error().Msgf("[AccrualService] responce error, status [%v] for order [%v]", response.StatusCode, *args.Order.Num)
			_, err := io.ReadAll(response.Body)
			if err != nil {
				ac.lg.Printf("Read response body error: %v", err)
				return err
			}
			return fmt.Errorf("[AccrualService] responce error, status [%v] for order [%v]", response.StatusCode, *args.Order.Num)
		}
	default:
		{
			ac.lg.Error().Msgf("[AccrualService] responce error, status [%v] for order [%v]", response.StatusCode, *args.Order.Num)
			_, err := io.ReadAll(response.Body)
			if err != nil {
				ac.lg.Printf("Read response body error: %v", err)
				return err
			}
			return fmt.Errorf("[AccrualService] responce error, status [%v] for order [%v]", response.StatusCode, *args.Order.Num)

		}
	}

	return nil
}
