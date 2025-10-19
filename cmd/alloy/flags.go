package main

import (
	"flag"
	"os"
)

type Flags struct {
	Port   string
	Dir    string
	Output string
}

func ParseFlags(args []string) (*Flags, []string, error) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	flags := &Flags{}
	fs.StringVar(&flags.Port, "port", "8080", "Port for dev server")
	fs.StringVar(&flags.Dir, "dir", ".", "Project directory")
	fs.StringVar(&flags.Output, "output", "", "Output path for build")

	err := fs.Parse(args)
	if err != nil {
		return nil, nil, err
	}

	return flags, fs.Args(), nil
}
