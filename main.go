package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	"golang.org/x/xerrors"

	"github.com/ghodss/yaml"
	"github.com/konoui/boltdb-exporter/pkg/exporter"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type config struct {
	filename        string
	outputFormat    string
	outputDirectory string
	marshaler       func(interface{}) ([]byte, error)
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
	fs.StringVar(&cfg.outputDirectory, "out", "", "directory output file will saved")
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

	b, err := exporter.Export(cfg.filename, cfg.marshaler)
	if err != nil {
		return err
	}
	for bucketName, byteData := range b {
		err := os.MkdirAll(cfg.outputDirectory, 0777)
		if err != nil {
			return xerrors.Errorf("Cannot create %s: %w", cfg.outputDirectory, err)
		}
		bucketFile := fmt.Sprintf("%s.json", bucketName)

		filePath := path.Join(cfg.outputDirectory, bucketFile)

		f, err := os.Create(filePath)
		if err != nil {
			return xerrors.Errorf("unable to open %s: %w", filePath, err)
		}

		if _, err = f.Write(byteData); err != nil {
			return xerrors.Errorf("failed to save a file: %w", err)
		}
		//fmt.Fprintln(os.Stdout, string(byteData))
	}
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
