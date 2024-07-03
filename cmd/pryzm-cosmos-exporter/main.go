package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pfc-developer/cosmos-exporter/pkg/exporter"
)

var (
	config exporter.ServiceConfig
	log    = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
)

var (
	Oracle bool
	LCD    string
)

var rootCmd = &cobra.Command{
	Use:  "cosmos-exporter",
	Long: "Scrape the data about the validators set, specific validators or wallets in the Cosmos network.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if config.ConfigPath == "" {
			config.SetBechPrefixes(cmd)

			return nil
		}

		viper.SetConfigFile(config.ConfigPath)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				log.Info().Err(err).Msg("Error reading config file")
				return err
			}
		}

		// Credits to https://carolynvanslyck.com/blog/2020/08/sting-of-the-viper/
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if !f.Changed && viper.IsSet(f.Name) {
				val := viper.Get(f.Name)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					log.Fatal().Err(err).Msg("Could not set flag")
				}
			}
		})
		config.SetBechPrefixes(cmd)

		return nil
	},
	Run: Execute,
}

func Execute(_ *cobra.Command, _ []string) {
	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not parse log level")
	}

	if config.JSONOutput {
		log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	zerolog.SetGlobalLevel(logLevel)

	config.LogConfig(log.Info()).
		Bool("oracle", Oracle).
		Str("lcd", LCD).
		Msg("Started with following parameters")

	sdkconfig := sdk.GetConfig()
	sdkconfig.SetBech32PrefixForAccount(config.AccountPrefix, config.AccountPubkeyPrefix)
	sdkconfig.SetBech32PrefixForValidator(config.ValidatorPrefix, config.ValidatorPubkeyPrefix)
	sdkconfig.SetBech32PrefixForConsensusNode(config.ConsensusNodePrefix, config.ConsensusNodePubkeyPrefix)
	sdkconfig.Seal()

	s := &exporter.Service{}

	s.Log = log
	err = s.Connect(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to service")
	}
	defer func(service *exporter.Service) {
		err := service.Close()
		if err != nil {
			s.Log.Fatal().Err(err).Msg("Could not close service client")
		}
	}(s)

	s.SetChainID(&config)
	s.SetDenom(&config)

	s.Params = config.Params
	s.Wallets = config.Wallets
	s.Validators = config.Validators
	s.Proposals = config.Proposals
	s.Oracle = config.Oracle
	s.Params = config.Params
	s.Upgrades = config.Upgrades
	s.Config = &config

	if config.SingleReq {
		log.Info().Msg("Starting Single Mode")
		http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { InjSingleHandler(w, r, s) })
	}
	http.HandleFunc("/metrics/wallet", s.WalletHandler)
	http.HandleFunc("/metrics/validator", s.ValidatorHandler)
	http.HandleFunc("/metrics/validators", s.ValidatorsHandler)
	http.HandleFunc("/metrics/params", s.ParamsHandler)
	http.HandleFunc("/metrics/general", s.GeneralHandler)

	http.HandleFunc("/metrics/delegator", s.DelegatorHandler)
	http.HandleFunc("/metrics/proposals", s.ProposalsHandler)
	http.HandleFunc("/metrics/upgrade", s.UpgradeHandler)
	if config.Prefix == "pryzm" {
		http.HandleFunc("/metrics/pryzm", func(w http.ResponseWriter, r *http.Request) { PryzmMetricHandler(w, r, s) })
	}

	log.Info().Str("address", config.ListenAddress).Msg("Listening")
	err = http.ListenAndServe(config.ListenAddress, nil) // #nosec
	if err != nil {
		log.Error().Err(err).Msg("could not start application")
		return
	}
}

func main() {
	config.SetCommonParameters(rootCmd)
	rootCmd.PersistentFlags().BoolVar(&Oracle, "oracle", false, "serve pryzm oracle info in the single call to /metrics")
	rootCmd.PersistentFlags().StringVar(&LCD, "lcd", "http://localhost:1317", "LCD endpoint")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
