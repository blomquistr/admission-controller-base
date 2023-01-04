package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var (
	port int
)

func setPort() {
	var err error = nil
	port, err = strconv.Atoi(os.Getenv("SERVER_PORT"))

	if err != nil {
		panic(err)
	}
}

func Run() {
	setPort()

	/* Here we define the endpoints we serve; we'll need one for each admission
	controller. Notice how they call the wrapper functions defined immediately above
	the main function. This makes unit test coverage easier, and also leads to more
	readable code. */
	http.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("pong")) })
	/* The last endpoint we want to define is our readiness probe - it's simple,
	so we'll just use an IIFE (immediately instantiated function expression) */
	http.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })
	
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
