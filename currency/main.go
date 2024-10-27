package main

import (
	"net"
	"os"

	"github.com/hashicorp/go-hclog"
	protos "github.com/hnsia/go-nic/currency/protos/currency/currency"
	"github.com/hnsia/go-nic/currency/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := hclog.Default()

	gs := grpc.NewServer()
	cs := server.NewCurrency(logger)

	protos.RegisterCurrencyServer(gs, cs)

	reflection.Register(gs) // Should disable this in production

	l, err := net.Listen("tcp", ":9092")
	if err != nil {
		logger.Error("Unable to listen", "error", err)
		os.Exit(1)
	}

	gs.Serve(l)
}
