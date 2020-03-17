package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AdamSLevy/flagbind"
)

// Flags defines all command line flags used by the program.
//
// The `flag` struct tag is parsed by my flagbind package to populate a
// flag.FlagSet.
//
// The tag value: <flag name>;<default value>;<usage>
// Flag names default to their kebab-case equivalent.
type Flags struct {
	Start string `flag:";;Start crawling the URL"`
	Stop  string `flag:";;Stop crawling the URL"`
	List  bool   `flag:";;List current site tree for all URLs"`

	action string

	Debug bool `flag:";;Print additional debug information"`
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
		return err
	}

	switch flags.action {
	case "list":
		return list()
	case "start":
		return start(flags.Start)
	case "stop":
		return stop(flags.Stop)
	default:
		fs.Usage()
		return flag.ErrHelp
	}
}
