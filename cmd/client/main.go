package main

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/api"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/app"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/commands"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/storage"
	"os"
	"path/filepath"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	serverAddr := flag.String("s", "localhost:3200", "gRPC server address")
	flag.Parse()

	home, _ := os.UserHomeDir()
	storePath := filepath.Join(home, ".gophkeeper.json")

	store, err := storage.Load(storePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load storage: %v\n", err)
		os.Exit(1)
	}

	client, err := api.NewClient(*serverAddr, store.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create client: %v\n", err)
		os.Exit(1)
	}

	application := &app.App{
		Client:       client,
		Storage:      store,
		ReadPassword: commands.ReadPassword,
	}

	var rootCmd = &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a secure password manager",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GophKeeper Client\nVersion: %s\nBuild Date: %s\n", buildVersion, buildDate)
		},
	})

	rootCmd.AddCommand(commands.RegisterCmd(application))
	rootCmd.AddCommand(commands.LoginCmd(application))
	rootCmd.AddCommand(commands.AddCmd(application))
	rootCmd.AddCommand(commands.ListCmd(application))
	rootCmd.AddCommand(commands.GetCmd(application))
	rootCmd.AddCommand(commands.DeleteCmd(application))
	rootCmd.AddCommand(commands.SyncCmd(application))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
