package cmd

import (
	"fmt"
	"os"

	"github.com/equinor/seismic-cloud/api/service"

	"github.com/equinor/seismic-cloud/api/config"
	"github.com/equinor/seismic-cloud/api/server"
	"github.com/equinor/seismic-cloud/api/service/store"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve seismic cloud provider",
	Long:  `serve seismic cloud provider.`,
	Run:   runServe,
}

func runServe(cmd *cobra.Command, args []string) {

	if viper.ConfigFileUsed() == "" {
		jww.ERROR.Println("Config from environment variables")
	} else {
		jww.INFO.Println("Using config file:", viper.ConfigFileUsed())
	}

	opts := make([]server.HttpServerOption, 0)

	if config.UseAuth() {
		opts = append(opts,
			server.WithOAuth2(config.AuthServer(), config.ResourceID(), config.Issuer()))
	}

	if len(config.ManifestStoragePath()) > 0 {
		opts = append(opts,
			server.WithManifestStore(&store.ManifestFileStore{
				BasePath: config.ManifestStoragePath()}))
	}

	var st service.Stitcher
	if len(config.StitchAddr()) > 0 {
		var err error
		st, err = service.NewTCPStitch(config.StitchAddr())
		if err != nil {
			jww.ERROR.Println("Stitch tcp error", err)
			os.Exit(1)
		}
	} else if len(config.StitchCmd()) > 0 {
		var err error
		st, err = service.NewExecStitch(config.StitchCmd(), config.Profiling())
		if err != nil {
			jww.ERROR.Println("Stitch exec error", err)
			os.Exit(1)
		}

	}
	opts = append(
		opts,
		server.WithStitcher(st))

	if len(config.HostAddr()) > 0 {
		opts = append(
			opts,
			server.WithHostAddr(config.HostAddr()))
	}

	if config.HttpOnly() {
		opts = append(
			opts,
			server.WithHttpOnly())
	}

	if config.UseLetsEncrypt() {
		opts = append(
			opts,
			server.WithLetsEncrypt(config.DomainList(), config.DomainMail()))
	}

	if config.UseTLS() {
		opts = append(
			opts,
			server.WithTLS(config.CertFile(), config.KeyFile()))
	}

	if config.Profiling() {
		opts = append(
			opts,
			server.WithProfiling())
	}

	if config.Swagger() {
		opts = append(
			opts,
			server.WithSwagger())
	}

	hs, err := server.NewHttpServer(opts...)

	if err != nil {
		fmt.Println("Error starting http server: ", err)
		os.Exit(1)
	}
	err = hs.Serve()

	if err != nil {
		fmt.Println("Error starting http server: ", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
