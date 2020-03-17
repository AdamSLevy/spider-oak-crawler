package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/AdamSLevy/flagbind"
	"github.com/PuerkitoBio/fetchbot"
)

// Flags defines all command line flags used by the program.
//
// The `flag` struct tag is parsed by my flagbind package to populate a
// flag.FlagSet.
//
// The tag value: <flag name>;<default value>;<usage>;<options>
// Flag names default to their kebab-case equivalent.
type Flags struct {
	Port int `flag:";9090;Port to serve gRPC API"`

	TLS      bool
	CertFile string
	KeyFile  string

	// flagbind exposes all of the options of the Fetcher as flags.
	// I'm working on a way to better control flags of nested structs with
	// flagbind.
	Fetcher fetchbot.Fetcher `flag:";;;flatten"`

	HTTPTimeout time.Duration `flag:";5s;HTTP Request Timeout"`

	// TODO: URL Normalization Rule Flags

	// TODO: add a debug logger
	Debug bool `flag:";;Print additional debug information"`
}

func main() {
	if err := _main(os.Args); err != nil {
		if err != flag.ErrHelp { // Don't print "flag: help requested"
			fmt.Println(err)
		}
		os.Exit(1)
	}
}
func _main(args []string) error {

	// Init & parse flags...
	var flags Flags
	//flags.Fetcher = fetchbot.New(nil) // TODO: write handler
	fs := flag.NewFlagSet("crawld", flag.ContinueOnError)
	if err := flagbind.Bind(fs, &flags); err != nil {
		// Should never happen with a well-formed flags struct.
		panic(err)
	}
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	// Listen for an Interrupt and cancel everything if it occurs.
	ctx, cancel := context.WithCancel(context.Background())
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	go func() {
		<-sigint
		cancel()
	}()

	g, _, err := StartServer(ctx, &flags)
	if err != nil {
		return err
	}
	fmt.Println("Waiting for crawl requests...")

	return g.Wait()
}
