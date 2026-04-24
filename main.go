package main

import (
	"flag"
	"log"

	dummy "git.happydns.org/checker-dummy/checker"
	"git.happydns.org/checker-sdk-go/checker/server"
)

// Version is the standalone binary's version. It defaults to "custom-build"
// and is meant to be overridden by the CI at link time:
//
//	go build -ldflags "-X main.Version=1.2.3" .
var Version = "custom-build"

var listenAddr = flag.String("listen", ":8080", "HTTP listen address")

func main() {
	flag.Parse()

	// Propagate the binary version to the checker package so it shows up in
	// CheckerDefinition.Version.
	dummy.Version = Version

	srv := server.New(dummy.Provider())
	if err := srv.ListenAndServe(*listenAddr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
