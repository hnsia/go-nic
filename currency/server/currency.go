package server

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hnsia/go-nic/currency/data"
	protos "github.com/hnsia/go-nic/currency/protos/currency/currency"
	"google.golang.org/grpc"
)

type Currency struct {
	rates *data.ExchangeRates
	log   hclog.Logger
	protos.UnimplementedCurrencyServer
}

func NewCurrency(r *data.ExchangeRates, l hclog.Logger) *Currency {
	return &Currency{r, l, protos.UnimplementedCurrencyServer{}}
}

func (c *Currency) GetRate(ctx context.Context, rr *protos.RateRequest) (*protos.RateResponse, error) {
	c.log.Info("Handle GetRate", "base", rr.GetBase(), "destination", rr.GetDestination())

	rate, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())
	if err != nil {
		return nil, err
	}

	return &protos.RateResponse{Rate: rate}, nil
}

func (c *Currency) SubscribeRates(src grpc.BidiStreamingServer[protos.RateRequest, protos.RateResponse]) error {

	go func() {
		for {
			rr, err := src.Recv()
			if err == io.EOF {
				c.log.Info("Client has closed connection")
				break
			}

			if err != nil {
				c.log.Error("Unable to read from client", "error", err)
				break
			}
			c.log.Info("Handle client request", "request", rr)
		}
	}()

	for {
		err := src.Send(&protos.RateResponse{Rate: 12.1})
		if err != nil {
			return err
		}

		time.Sleep(5 * time.Second)
	}
}
