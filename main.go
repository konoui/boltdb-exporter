package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/konoui/boltdb-exporter/pkg/exporter"
	"github.com/peterbourgon/ff/ffcli"
)

type config struct {
	filename     string
	outputFormat string
	marshaler    func(interface{}) ([]byte, error)
}

var (
	conf      config
	rootFlags = flag.NewFlagSet("boltdb-exporter", flag.ExitOnError)
	rootCmd   = &ffcli.Command{
		Usage:     "boltdb-exporter --db <db filename> [flags...]",
		ShortHelp: "expot/dump boltdb as json/yaml format",
		FlagSet:   rootFlags,
		Exec: func(args []string) error {
			return conf.run()
		},
	}
)

func (c *config) validate() error {
	if c.filename == "" {
		return fmt.Errorf("database file option is not specified ")
	}

	if _, err := os.Stat(c.filename); os.IsNotExist(err) {
		return fmt.Errorf("databse file %s does not exist", c.filename)
	}

	switch c.outputFormat {
	case "json":
		c.marshaler = func(i interface{}) ([]byte, error) {
			return json.MarshalIndent(i, "", "  ")
		}
	case "yaml", "yml":
		c.marshaler = yaml.Marshal
	default:
		return fmt.Errorf("%s is an invalid output format", c.outputFormat)
	}
	return nil
}

func (c *config) run() error {
	if err := c.validate(); err != nil {
		return err
	}

	b, err := exporter.Export(c.filename, c.marshaler)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(b))
	return nil
}

func init() {
	rootFlags.StringVar(&conf.outputFormat, "format", "json", "support json/yaml")
	rootFlags.StringVar(&conf.filename, "db", "", "database filename")
}

func main() {
	if err := rootCmd.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		fmt.Fprintf(os.Stderr, "Usage: %s\n", rootCmd.Usage)
		os.Exit(1)
	}
}
