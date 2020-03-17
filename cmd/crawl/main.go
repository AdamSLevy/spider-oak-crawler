package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/AdamSLevy/flagbind"
	pb "github.com/AdamSLevy/spider-oak-crawler/internal/crawl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Flags defines all command line flags used by the program.
//
// The `flag` struct tag is parsed by my flagbind package to populate a
// flag.FlagSet.
//
// The tag value: <flag name>;<default value>;<usage>
// Flag names default to their kebab-case equivalent.
type Flags struct {
	Start  string `flag:";;Start crawling the URL"`
	Stop   string `flag:";;Stop crawling the URL"`
	List   bool   `flag:";;List current site tree for all URLs"`
	action string // Set by Validate(), may be "start", "stop", or "list"

	Debug bool `flag:";;Print additional debug information"`

	// API Settings
	Server string `flag:";localhost:9090;Hostname:port of gRPC API"`

	// TLS Settings
	TLS                bool
	CAFile             string
	ServerHostOverride string
}

// Validate ensures that the provided flags were well-formed.
//
// -start, -stop, -list are mutually exclusive
func (f *Flags) Validate() error {

	// Count # of mutually exclusive flags...
	var set int
	// Zero-value is not a perfect metric for flag presence, but good
	// enough.
	if f.Start != "" {
		set++
		f.action = "start"
	}
	if f.Stop != "" {
		set++
		f.action = "stop"
	}
	if f.List {
		set++
		f.action = "list"
	}

	if set == 0 {
		return flag.ErrHelp
	}

	if set > 1 {
		return fmt.Errorf("-start, -stop, and -list are mutually exclusive")
	}

	return nil
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
	fs := flag.NewFlagSet("crawl", flag.ContinueOnError)
	if err := flagbind.Bind(fs, &flags); err != nil {
		// Should never happen with a well-formed flags struct.
		panic(err)
	}
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if err := flags.Validate(); err != nil {
		if err == flag.ErrHelp {
			fs.Usage()
		}
		return err
	}

	// Set up gRPC client
	var opts []grpc.DialOption
	if flags.TLS {
		creds, err :=
			credentials.NewClientTLSFromFile(
				flags.CAFile, flags.ServerHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(flags.Server, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewCrawlClient(conn)

	switch flags.action {
	case "list":
		return list(client)
	case "start":
		return start(client, flags.Start)
	case "stop":
		return stop(client, flags.Stop)
	default:
		panic(fmt.Errorf("unknown flags.action: %q", flags.action))
	}
}
