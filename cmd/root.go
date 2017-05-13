package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/cmd/consumer"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/cmd/publisher"
)

const defaultConfig = `# RDSS Archivematica Channel Adapter

[logging]
level = "INFO"

[amclient]
# URL of the Archivematica Dashboard
# url = "https://archivematica.internal:9000"
# user = "demo"
# key = "eid3Aitheijoo1ohce2pho4eiDei0lah"

[s3]
force_path_style = false
insecure_skip_verify = false
# endpoint = "https://127.0.0.1:4567"
# access_key = "..."
# secret_key = "..."
# region = "..."

[consumer]
archivematica_transfer_deposit_dir = "/var/archivematica/sharedDirectory/watchedDirectories/activeTransfers/standardTransfer"

[publisher]
listen = "0.0.0.0:8000"
tls = false
# tls_cert_file = "/foo.crt"
# tls_key_file = "/foo.key"

[broker]
backend = "kinesis"

[broker.queues]
main = "main"
invalid = "invalid"
error = "error"

[broker.kinesis]
# This adapter uses aws-sdk-go. The credentials must be defined using the
# canonical environment variables, read more at https://goo.gl/xsWyS9.

tls = true

# Custom endpoints for Kinesis and DynamoDB, useful in local environments.
#endpoint = "https://127.0.0.1:4567"
#dynamodb_endpoint = "https://127.0.0.1:4568"
`

var (
	cfgFile string

	logger = log.StandardLogger()
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "rdss-archivematica-channel-adapter",
	Short: "RDSS Archivematica Channel Adapter",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, setupLogger)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rdss-archivematica-channel-adapter.toml)")

	RootCmd.AddCommand(consumer.Command(logger.WithFields(log.Fields{"cmd": "consumer"})))
	RootCmd.AddCommand(publisher.Command(logger.WithFields(log.Fields{"cmd": "publisher"})))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // Enable ability to specify config file via flag.
		viper.SetConfigFile(cfgFile)
	}

	// Name of config file (without extension).
	viper.SetConfigName(".rdss-archivematica-channel-adapter")

	// Adding home directory as first search path.
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("/etc/archivematica")

	// Read in environment variables that match.
	viper.SetEnvPrefix("RDSS_ARCHIVEMATICA_ADAPTER")
	viper.AutomaticEnv()

	// Set our preferred configuration format.
	viper.SetConfigType("toml")

	// Read our default configuration.
	if err := viper.ReadConfig(strings.NewReader(defaultConfig)); err != nil {
		logger.Fatalln("Cannot read configuration file:", err)
	}

	// If a config file is found, read it in.
	if err := viper.MergeInConfig(); err == nil {
		logger.Infoln("Using config file:", viper.ConfigFileUsed())
	}
}

func setupLogger() {
	var (
		input = viper.GetString("logging.level")
		level = log.InfoLevel
	)

	l, err := log.ParseLevel(input)
	if err != nil {
		logger.Errorln("Not a valid logging level:", input)
	} else {
		level = l
	}

	log.SetLevel(level)
}
