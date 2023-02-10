package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/blomquistr/admission-controller-base/internal/server"
)

var (
	// Used for flags
	cfgFile string
	certFile string
	keyFile string
	message string
	port int

	// the actual command
	rootCmd = &cobra.Command{
		Use: "server",
		Short: "A mutating admission controller webhook server scaffold",
		Long: `A scaffolding for a mutating admission controller server that
				is serving as a project for learning some different Go libraries`,
		Run: run,
	}
)

// the exported command that the program runs
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// init is the 
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "a configuration file to use to start the server (defaults to $HOME/.webhook/server.yml)")
	rootCmd.PersistentFlags().StringVar(&certFile, "cert-file", "", "the path to a valid TLS certificate file")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key-file", "", "the path to a valid TLS certificate key file for the provided cert file")
	rootCmd.PersistentFlags().StringVar(&message, "message", "Hello, World!", "A message to pass to the application's test endpoint to validate that it is working")
	rootCmd.PersistentFlags().IntVar(&port, "port", 5001, "A port to run the server on (defaults to 5001)")

}

func initConfig() {

}

func run(cmd *cobra.Command, args []string) {
	server.Run()
}