package publisher

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/publisher"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/publisher/pb"

	// Backend implementations
	_ "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/kinesis"

	// Serve runtime profiling data via HTTP
	_ "net/http/pprof"
)

var cmd = &cobra.Command{
	Use:   "publisher",
	Short: "Publisher server (Archivematica » RDSS)",
	Run:   server,
}

var logger log.FieldLogger

func Command(l log.FieldLogger) *cobra.Command {
	logger = l
	return cmd
}

func server(cmd *cobra.Command, args []string) {
	logger.Info("Hello!")
	defer logger.Info("Bye!")

	go func() {
		logger.Errorln(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	lis, err := net.Listen("tcp", viper.GetString("publisher.listen"))
	if err != nil {
		logger.Fatalln(err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(unaryInterceptor()),
	}
	if viper.GetBool("publisher.tls") {
		creds, err := credentials.NewServerTLSFromFile(viper.GetString("publisher.tls-cert-file"), viper.GetString("publisher.tls-key-file"))
		if err != nil {
			logger.Fatalln(err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(opts...)
	bo, err := makeBroker()
	if err != nil {
		logger.Fatalln(err)
	}
	rdssServer := publisher.MakeRdssServer(bo, logger)
	pb.RegisterRdssServer(grpcServer, rdssServer)

	go func() {
		logger.Infof("gRCP server listening on %s", viper.GetString("publisher.listen"))
		grpcServer.Serve(lis)
	}()

	// Subscribe to SIGINT signals and wait
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	<-stopChan // Wait for SIGINT

	// Graceful shutdown with a timeout
	logger.Info("Shutting down server...")
	const timeout = time.Second * 3
	c := make(chan struct{})
	go func() {
		defer close(c)
		grpcServer.GracefulStop()
	}()
	select {
	case <-c:
		logger.Info("Server gracefully stopped!")
	case <-time.After(timeout):
		logger.Info("Server timedout when we tried to stop it.")
	}
}

func makeBroker() (*broker.Broker, error) {
	var (
		opts     []backend.DialOpts
		_        = viper.GetString("broker.kinesis.stream")
		endpoint = viper.GetString("broker.kinesis.endpoint")
	)
	if endpoint != "" {
		opts = append(opts, backend.WithKeyValue("endpoint", endpoint))
	}
	b, err := backend.Dial("kinesis", opts...)
	if err != nil {
		log.Fatalln(err)
	}
	return broker.New(b, logger, &broker.Config{QueueMain: "main", QueueError: "error", QueueInvalid: "invalid"})
}

func unaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Debug("gRPC request received")
		return handler(ctx, req)
	}
}
