package main

import (
	"log"
	"os"

	"github.com/canhlinh/hlsdl"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:          "hlsdl",
	RunE:         cmdF,
	SilenceUsage: true,
}

func main() {
	cmd.Flags().StringP("url", "u", "", "The manifest (m3u8) url")
	cmd.Flags().StringP("dir", "d", "./download", "The directory where the file will be stored")
	cmd.Flags().IntP("workers", "w", 2, "Number of workers to execute concurrent operations")
	cmd.SetArgs(os.Args[1:])

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func cmdF(command *cobra.Command, args []string) error {
	m3u8URL, err := command.Flags().GetString("url")
	if err != nil {
		return err
	}

	dir, err := command.Flags().GetString("dir")
	if err != nil {
		return err
	}

	workers, err := command.Flags().GetInt("workers")
	if err != nil {
		return err
	}

	hlsDL := hlsdl.New(m3u8URL, dir, workers)
	filepath, err := hlsDL.Download()
	if err != nil {
		return err
	}

	log.Println("Downloaded file to " + filepath)
	return nil
}
