package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/publisher/pb"
)

var (
	tls                bool
	caFile             string
	addr               string
	serverHostOverride string
)

// echoCmd represents the echo command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Example echo gRPC service CLI client",
	Run:   client,
}

func init() {
	RootCmd.AddCommand(clientCmd)

	clientCmd.Flags().BoolVarP(&tls, "tls", "", false, "TLS enabled")
	clientCmd.Flags().StringVarP(&caFile, "tls-ca-file", "", "", "CA root cert file")
	clientCmd.Flags().StringVarP(&addr, "addr", "", "127.0.0.1:8000", "Address of the gRPC server")
	clientCmd.Flags().StringVarP(&serverHostOverride, "server-host-override", "", "", "Server name used to verify the hostname returned by TLS handshake")
}

func client(cmd *cobra.Command, args []string) {
	var opts []grpc.DialOption
	if tls {
		sn := addr
		if serverHostOverride != "" {
			sn = serverHostOverride
		}
		var creds credentials.TransportCredentials
		if caFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(caFile, sn)
			if err != nil {
				log.Fatalf("Failed to create TLS credentials %v", err)
			}
		} else {
			creds = credentials.NewClientTLSFromCert(nil, sn)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewRdssClient(conn)

	log.Println("Request: Read Metadata")
	resp, err := client.MetadataRead(context.Background(), &pb.MetadataReadRequest{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Response: %v", resp)
}
