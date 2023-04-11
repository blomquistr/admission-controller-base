package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"

	"github.com/blomquistr/admission-controller-base/internal/server"
)

var (
	// Used for flags
	cfgFile      string
	certFile     string
	keyFile      string
	message      string
	port         int
	configPrefix string = "webhook"

	// the actual command
	rootCmd = &cobra.Command{
		Use:   "server",
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

	// flag config - the path to the config file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "a configuration file to use to start the server (defaults to $HOME/.webhook/server.yml)")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))                      // binds the PFlag to a Viper config value
	viper.BindEnv("config", fmt.Sprintf("%s_CONFIG_FILE_PATH", strings.ToUpper(configPrefix))) // binds an environment variable to the Viper config value
	viper.SetDefault("config", "$HOME/.admission-controller-base/config")                      // sets a default for the config value

	// flag cert-file, the path to the TLS cert
	rootCmd.PersistentFlags().StringVar(&certFile, "cert-file", "", "the path to a valid TLS certificate file")
	viper.BindPFlag("certFile", rootCmd.PersistentFlags().Lookup("cert-file"))
	viper.BindEnv("certFile", fmt.Sprintf("%s_CERT_FILE_PATH", strings.ToUpper(configPrefix)))
	viper.SetDefault("certFile", "")

	// flag key-file, the path to the TLS.key file
	rootCmd.PersistentFlags().StringVar(&keyFile, "key-file", "", "the path to a valid TLS certificate key file for the provided cert file")
	viper.BindPFlag("keyFile", rootCmd.PersistentFlags().Lookup("key-file"))
	viper.BindEnv("keyFile", fmt.Sprintf("%s_KEY_FILE_PATH", strings.ToUpper(configPrefix)))
	viper.SetDefault("keyFile", "")

	// flag port - the port the app should listen on in its environment
	rootCmd.PersistentFlags().IntVar(&port, "port", 5001, "A port to run the server on (defaults to 5001)")
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindEnv("port", fmt.Sprintf("%s_PORT", strings.ToUpper(configPrefix)))
	viper.SetDefault("port", 5001)

	// flag message - a test message for passing around the app, basically hello-world for configuration in Viper and Cobra
	rootCmd.PersistentFlags().StringVar(&message, "message", "Hello, World!", "A message to pass to the application's test endpoint to validate that it is working")
	viper.BindPFlag("message", rootCmd.PersistentFlags().Lookup("message"))
	viper.BindEnv("message", fmt.Sprintf("%s_MESSAGE", strings.ToUpper(configPrefix)))
	viper.SetDefault("message", "Hello World!")
}

func initConfig() {
	if cfgFile != "" {
		// use the config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// find the user's home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		cwd, err := os.Getwd()
		cobra.CheckErr(err)

		// search the home directory for a subdirectory named ".admission-controller-base"
		// and find a file in it called "config" with a file extension of "yaml" or "json"
		// to facilitate development, we will also search the current working directory; a
		// basic configuration example is included in the repository.
		viper.AddConfigPath(fmt.Sprintf("%s/.admission-controller", home))
		viper.AddConfigPath(fmt.Sprintf("%s/examples/config", cwd))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix(configPrefix)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// config file is not found, ignore the error
		} else {
			// config file was found, but malformed
			klog.Fatal(ok)
		}
	}

}

// This wrapper will actually run the server
func run(cmd *cobra.Command, args []string) {
	server.Run()
}
