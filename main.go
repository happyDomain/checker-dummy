package main

import (
	"flag"
	"log"

	dummy "git.happydns.org/happyDomain/checker-dummy/checker"
	sdk "git.happydns.org/happyDomain/sdk/checker"
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

	server := sdk.NewServer(dummy.Provider())
	if err := server.ListenAndServe(*listenAddr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
