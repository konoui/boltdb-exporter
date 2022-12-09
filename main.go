package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/konoui/boltdb-exporter/pkg/exporter"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type arrayFlag []string

type config struct {
	filename        string
	outputFormat    string
	bucketSelection arrayFlag
	marshaler       func(interface{}) ([]byte, error)
}

func (i *arrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *arrayFlag) String() string {
	return "..."
}

func newRootCmd() *ffcli.Command {
	cfg := new(config)
	fs := flag.NewFlagSet("boltdb-exporter", flag.ExitOnError)
	cfg.registerFlags(fs)

	return &ffcli.Command{
		Name:       "boltdb-exporter",
		ShortUsage: "boltdb-exporter --db <db filename> [flags...]",
		ShortHelp:  "expot/dump boltdb as json/yaml format",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			return cfg.run()
		},
	}
}

func (cfg *config) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.outputFormat, "format", "json", "support json/yaml")
	fs.StringVar(&cfg.filename, "db", "", "database filename")
	fs.Var(&cfg.bucketSelection, "bucket", "select root-level bucket to export (can be used multiple times)")
}

func (cfg *config) validate() error {
	if cfg.filename == "" {
		return fmt.Errorf("database file option is not specified ")
	}

	if _, err := os.Stat(cfg.filename); err != nil {
		return fmt.Errorf("databse file %s does not exist", cfg.filename)
	}

	switch cfg.outputFormat {
	case "json":
		cfg.marshaler = func(i interface{}) ([]byte, error) {
			return json.MarshalIndent(i, "", "  ")
		}
	case "yaml", "yml":
		cfg.marshaler = yaml.Marshal
	default:
		return fmt.Errorf("%s is an invalid output format", cfg.outputFormat)
	}
	return nil
}

func (cfg *config) run() error {
	if err := cfg.validate(); err != nil {
		return err
	}

	var bucketSelectionSet map[string]bool
	if len(cfg.bucketSelection) > 0 {
		bucketSelectionSet = make(map[string]bool)
		for _, bucket := range cfg.bucketSelection {
			bucketSelectionSet[bucket] = true
		}
	}

	b, err := exporter.Export(cfg.filename, cfg.marshaler, bucketSelectionSet)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(b))
	return nil
}

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		fmt.Fprintf(os.Stderr, "Usage: %s\n", rootCmd.ShortUsage)
		os.Exit(1)
	}
}
