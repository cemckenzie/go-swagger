package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-swagger/go-swagger/spec"
	flags "github.com/jessevdk/go-flags"
	graceful "github.com/tylerb/graceful"

	"github.com/go-swagger/go-swagger/examples/tutorials/todo-list/server-1/restapi/operations"
)

// This file was generated by the swagger tool.
// Make sure not to overwrite this file after you generated it because all your edits would be lost!
// It would only be overwritten if you explicitly specify --include-main for the generate all or support commands
//go:generate swagger generate server -t ../.. -A TodoList

var opts struct {
	Host string `long:"host" description:"the IP to listen on" default:"localhost" env:"HOST"`
	Port int    `long:"port" description:"the port to listen on for insecure connections, defaults to a random value" env:"PORT"`
}

func main() {
	swaggerSpec, err := spec.New(SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	parser := flags.NewParser(&opts, flags.Default)
	parser.ShortDescription = swaggerSpec.Spec().Info.Title
	parser.LongDescription = swaggerSpec.Spec().Info.Description

	api := operations.NewTodoListAPI(swaggerSpec)
	handler := configureAPI(api)

	for _, optsGroup := range api.CommandLineOptionsGroups {
		parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
	}

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	httpServer := &graceful.Server{Server: new(http.Server)}
	httpServer.Handler = handler

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.Port))
	if err != nil {
		api.ServerShutdown()
		log.Fatalln(err)
	}

	fmt.Printf("serving todo list at http://%s\n", listener.Addr())
	if err := httpServer.Serve(tcpKeepAliveListener{listener.(*net.TCPListener)}); err != nil {
		api.ServerShutdown()
		log.Fatalln(err)
	}

	go func() {

		<-httpServer.StopChan()

		api.ServerShutdown()
	}()
}

// tcpKeepAliveListener is copied from the stdlib net/http package

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
