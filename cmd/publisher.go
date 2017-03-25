package cmd

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/outbound"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/outbound/pb"

	// Backend implementations
	_ "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/kinesis"
)

var (
	serverInterface       string
	serverPort            int
	serverTLSEnabled      bool
	serverTLSCertFile     string
	serverTLSKeyFile      string
	serverKinesisStream   string
	serverKinesisEndpoint string
)

// publisherCmd represents the publisher command
var publisherCmd = &cobra.Command{
	Use:   "publisher",
	Short: "Outbound server (Archivematica » RDSS)",
	Run:   server,
}

func init() {
	RootCmd.AddCommand(publisherCmd)

	publisherCmd.Flags().StringVarP(&serverInterface, "bind", "b", "127.0.0.1", "interface to which the gRPC server will bind")
	publisherCmd.Flags().IntVarP(&serverPort, "port", "p", 8000, "port on which the gRPC server will listen")
	publisherCmd.Flags().BoolVarP(&serverTLSEnabled, "tls", "", false, "TLS enabled")
	publisherCmd.Flags().StringVarP(&serverTLSCertFile, "tls-cert-file", "", "", "TLS cert file")
	publisherCmd.Flags().StringVarP(&serverTLSKeyFile, "tls-key-file", "", "", "TLS key file")
	publisherCmd.Flags().StringVarP(&serverKinesisStream, "kinesis-stream", "", "", "Kinesis stream")
	publisherCmd.Flags().StringVarP(&serverKinesisEndpoint, "kinesis-endpoint", "", "", "Kinesis endpoint, e.g. https://127.0.0.1:4567")
}

func server(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	addr := net.JoinHostPort(serverInterface, strconv.Itoa(serverPort))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	var opts []grpc.ServerOption
	if serverTLSEnabled {
		creds, err := credentials.NewServerTLSFromFile(serverTLSCertFile, serverTLSCertFile)
		if err != nil {
			log.Fatalln(err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRdssServer(grpcServer, outbound.MakeRdssServer(getBroker()))

	go func() {
		log.Println("gRPC server listening on", addr)
		grpcServer.Serve(lis)
	}()

	// Subscribe to SIGINT signals and wait
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	<-stopChan // Wait for SIGINT

	// Graceful shutdown with a timeout
	log.Println("Shutting down server...")
	const timeout = time.Second * 3
	c := make(chan struct{})
	go func() {
		defer close(c)
		grpcServer.GracefulStop()
	}()
	select {
	case <-c:
		log.Println("Server gracefully stopped!")
	case <-time.After(timeout):
		log.Println("Server timedout when we tried to stop it.")
	}
}

func getBroker() broker.Backend {
	var opts []broker.DialOpts

	if serverKinesisStream != "" {
		opts = append(opts, broker.WithKeyValue("stream", serverKinesisStream))
	}

	if serverKinesisEndpoint != "" {
		opts = append(opts, broker.WithKeyValue("endpoint", serverKinesisEndpoint))
	}

	b, err := broker.Dial("kinesis", opts...)
	if err != nil {
		log.Fatalln(err)
	}
	return b
}
