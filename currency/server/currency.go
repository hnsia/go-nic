package server

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hnsia/go-nic/currency/data"
	protos "github.com/hnsia/go-nic/currency/protos/currency/currency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Currency struct {
	rates         *data.ExchangeRates
	log           hclog.Logger
	subscriptions map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest
	protos.UnimplementedCurrencyServer
}

func NewCurrency(r *data.ExchangeRates, l hclog.Logger) *Currency {
	subscriptions := make(map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest)

	c := &Currency{r, l, subscriptions, protos.UnimplementedCurrencyServer{}}
	go c.handleUpdates()

	return c
}

func (c *Currency) handleUpdates() {
	ru := c.rates.MonitorRates(5 * time.Second)
	for range ru {
		c.log.Info("Got updated rates")

		// loop over subscribed clients
		for k, v := range c.subscriptions {

			// loop over subscribed rates
			for _, rr := range v {
				r, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())
				if err != nil {
					c.log.Error("Unable to get updated rate", "base", rr.GetBase().String(), "destination", rr.GetDestination().String())
				}

				err = k.Send(&protos.RateResponse{Base: rr.Base, Destination: rr.Destination, Rate: r})
				if err != nil {
					c.log.Error("Unable to send updated rate", "base", rr.GetBase().String(), "destination", rr.GetDestination().String())
				}
			}
		}
	}
}

func (c *Currency) GetRate(ctx context.Context, rr *protos.RateRequest) (*protos.RateResponse, error) {
	c.log.Info("Handle GetRate", "base", rr.GetBase(), "destination", rr.GetDestination())

	if rr.Base == rr.Destination {
		err := status.Newf(
			codes.InvalidArgument,
			"Base currency %s can not be the same as the destination currency %s",
			rr.Base.String(),
			rr.Destination.String(),
		)

		err, wde := err.WithDetails(rr)
		if wde != nil {
			return nil, wde
		}

		return nil, err.Err()
	}

	rate, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())
	if err != nil {
		return nil, err
	}

	return &protos.RateResponse{Base: rr.Base, Destination: rr.Destination, Rate: rate}, nil
}

// SubscribeRates implements the gRPC bidirectional streaming method for the server
func (c *Currency) SubscribeRates(src grpc.BidiStreamingServer[protos.RateRequest, protos.StreamingRateResponse]) error {

	// handle client messages
	for {
		rr, err := src.Recv() // Recv is a blocking method which returns on client data
		// io.EOF signals that the client has closed the connection
		if err == io.EOF {
			c.log.Info("Client has closed connection")
			break
		}

		// any other error means the transport between the server and client is unavailable
		if err != nil {
			c.log.Error("Unable to read from client", "error", err)
			return err
		}

		c.log.Info("Handle client request", "request", rr)
		rrs, ok := c.subscriptions[src]
		if !ok {
			rrs = []*protos.RateRequest{}
		}

		// check that subscription does not exists
		var validationError *status.Status
		for _, v := range rrs {
			if v.Base == rr.Base && v.Destination == rr.Destination {
				// subscription exists return errors
				validationError := status.Newf(
					codes.AlreadyExists,
					"Unable to subscribe for currency as subscription already exists",
				)

				// add the original request as metadata
				validationError, err = validationError.WithDetails(rr)
				if err != nil {
					c.log.Error("Unable to add metadata to error", "error", err)
					break
				}

				break
			}
		}

		// if validation error return, error and continue
		if validationError != nil {
			src.Send(
				&protos.StreamingRateResponse{
					Message: &protos.StreamingRateResponse_Error{
						Error: validationError.Proto(),
					},
				},
			)
			continue
		}

		rrs = append(rrs, rr)
		c.subscriptions[src] = rrs
	}

	return nil
}

// {
// 	"Base": "JPY",
// 	"Destination": "USD"
// }
