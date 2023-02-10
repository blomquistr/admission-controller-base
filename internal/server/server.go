package server

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"

	"github.com/blomquistr/admission-controller-base/internal/webhook"
)

var (
	config IConfig
)

func LoadTLSCerts(certFile string, keyFile string) *tls.Config {
	sCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		klog.Fatal(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}
}

func readyzHandler(w http.ResponseWriter, req *http.Request) {
	klog.Info("Handling readiness probe")
	w.Write([]byte("ok"))
}

func pingHandler(w http.ResponseWriter, req *http.Request) {
	klog.Info("Handling a ping")
	w.Write([]byte("pong"))
}

func serveMessage(w http.ResponseWriter, req *http.Request) {
	klog.Info("Reading config message and returning it")
	w.Write([]byte(config.getMessage()))
}

func serveNoExternalIpLoadBalancers(w http.ResponseWriter, req *http.Request) {
	klog.Info("Checking load balancers for external IP address")
	serve(w, req, webhook.NewDelegateToV1AdmitHandler(webhook.NoExternalIpLoadBalancers))
}

func serveNoDefaultNamespace(w http.ResponseWriter, req *http.Request) {
	klog.Info("Checking resource to exclude the Default namespace")
	serve(w, req, webhook.NewDelegateToV1AdmitHandler(webhook.AdmitNoDefault))
}

func Run() {
	config = newConfig()

	// message = "Hello World!"

	/* Here we define the endpoints we serve; we'll need one for each admission
	controller. Notice how they call the wrapper functions defined immediately above
	the main function. This makes unit test coverage easier, and also leads to more
	readable code. */
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/message", serveMessage)

	// the handlers listed below this are the handlers that would be registered as
	// a mutating webhook or validating webhook with a cluster - just figuring out
	// how to import them is an adventure.
	http.HandleFunc("/no-external-ip-load-balancers", serveNoExternalIpLoadBalancers)
	http.HandleFunc("/no-default-namespace", serveNoDefaultNamespace)

	/* The last endpoint we want to define is our readiness probe - it's simple,
	so we'll just use an IIFE (immediately instantiated function expression) */
	http.HandleFunc("/readyz", readyzHandler)

	var err error // not sure if this is technically required, or if go can handle just defining err in the if/else
	if config.getCertFile() != "" && config.getKeyFile() != "" {
		klog.Info("Certificate infromation identified, serving with TLS enabled...")
		// future state: serve with TLS
		server := &http.Server{
			Addr: fmt.Sprintf(":%d", config.getPort()),
			TLSConfig: LoadTLSCerts(config.getCertFile(), config.getKeyFile()),
		}
		err = server.ListenAndServeTLS("", "")
	} else {
		klog.Warning("No certificate data for TLS provided, falling back to serving unsecured endpoints...")
		klog.Warningf("Received path [%s] to cert file", config.getCertFile())
		klog.Warningf("Received path [%s] to key file", config.getKeyFile())
		server := &http.Server{
			Addr: fmt.Sprintf(":%d", config.getPort()),
		}
		err = server.ListenAndServe()
	}
	if err != nil {
		panic(err)
	}
}
