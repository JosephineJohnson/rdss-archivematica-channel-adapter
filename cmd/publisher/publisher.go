package publisher

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
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

	// Create the broker we're going to publish to
	bo, err := makeBroker()
	if err != nil {
		logger.Fatalln(err)
	}
	// Create an RDSS server using the broker
	rdssServer := publisher.MakeRdssServer(bo, logger)
	// Register RPC server with RDSS via protobuf
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRdssServer(grpcServer, rdssServer)

	go func() {
		logger.Infof("gRCP server listening on %s", viper.GetString("publisher.listen"))
		grpcServer.Serve(lis)
	}()

	// Subscribe to signals and wait
	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan // Block until a signal is received

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
		qM       = viper.GetString("broker.queues.main")
		qI       = viper.GetString("broker.queues.invalid")
		qE       = viper.GetString("broker.queues.error")
	)
	// Set to use given endpoint if given
	if endpoint != "" {
		opts = append(opts, backend.WithKeyValue("endpoint", endpoint))
	}

	b, err := backend.Dial("kinesis", opts...)
	if err != nil {
		log.Fatalln(err)
	}
	return broker.New(b, logger, &broker.Config{
		QueueMain:    qM,
		QueueError:   qE,
		QueueInvalid: qI,
		RepositoryConfig: &broker.RepositoryConfig{
			Backend: "dynamodb",
		},
	})
}

func unaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Debug("gRPC request received")
		return handler(ctx, req)
	}
}
