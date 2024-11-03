package main

import (
	"net"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hnsia/go-nic/currency/data"
	protos "github.com/hnsia/go-nic/currency/protos/currency/currency"
	"github.com/hnsia/go-nic/currency/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := hclog.Default()

	rates, err := data.NewRates(logger)
	if err != nil {
		logger.Error("Unable to generate rates", "error", err)
		os.Exit(1)
	}

	// create a new gRPC server, use WithInsecure to allow http connections
	gs := grpc.NewServer()

	// create an instance of the currency server
	cs := server.NewCurrency(rates, logger)

	// register the currency server
	protos.RegisterCurrencyServer(gs, cs)

	// register the reflection service which allows clients to determine the methods for this gRPC service
	reflection.Register(gs) // Should disable this in production

	// create a TCP socket for inbound server connections
	l, err := net.Listen("tcp", ":9092")
	if err != nil {
		logger.Error("Unable to listen", "error", err)
		os.Exit(1)
	}

	// listen for requests
	gs.Serve(l)
}
