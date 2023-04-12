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

func printConfig() {
	klog.Info("Printing configuration...")
	klog.Infof("CertFile: [%s]", config.getCertFile())
	klog.Infof("KeyFile: [%s]", config.getKeyFile())
	klog.Infof("Message: [%s]", config.getMessage())
	klog.Infof("HTTP Server Port: [%d]", config.getHttpPort())
	klog.Infof("HTTPS Server Port: [%d]", config.getHttpsPort())
}

func Run() {
	config = newConfig()

	printConfig()

	// wrapipng this in an if statement rather than checking on each call to klog
	// for brevity's sake and because why perform those extra checks?
	// if klog.V(2).Enabled() {
	// 	printConfig()
	// }

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

	if config.getCertFile() != "" && config.getKeyFile() != "" {
		/* In order to serve both readiness/liveness probes over HTTP and the mutating
		admission controller over https, we need to start the HTTP server in a goroutine
		and then the HTTPS server in the main process. */
		klog.Info("Certificate infromation identified, serving with TLS enabled...")
		go func() {
			http_server := &http.Server{
				Addr: fmt.Sprintf(":%d", config.getHttpPort()),
			}
			http_err := http_server.ListenAndServe()
			if http_err != nil {
				klog.Fatalf("HTTP Web Server Error: [%s]", http_err.Error())
			}
		}()
		// Now that the HTTP server is running, we can start the HTTPS server
		https_server := &http.Server{
			Addr:      fmt.Sprintf(":%d", config.getHttpsPort()),
			TLSConfig: LoadTLSCerts(config.getCertFile(), config.getKeyFile()),
		}
		https_err := https_server.ListenAndServeTLS("", "")
		if https_err != nil {
			panic(fmt.Sprintf("HTTPS Web Server Error: [%s]", https_err.Error()))
		}
	} else {
		/* This is a purely debugging path - if you're running locally, you might not
		bother serving HTTPS. To support this, we have an http-only path defined here. */
		klog.Warning("No certificate data for TLS provided, falling back to serving unsecured endpoints...")
		klog.Warningf("Received path [%s] to cert file", config.getCertFile()) // in case you _expected_ to see a cert file
		klog.Warningf("Received path [%s] to key file", config.getKeyFile())   // in case you _expected_ to see a key file
		server := &http.Server{
			Addr: fmt.Sprintf(":%d", config.getHttpPort()),
		}
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}
}
