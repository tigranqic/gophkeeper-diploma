// Package commands defines all CLI sub-commands for the GophKeeper client.
package commands

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/app"
	"github.com/tigranqic/gophkeeper-diploma/pkg/crypto"
	"golang.org/x/term"
	"os"
)

// ReadPassword reads a password from the terminal without echoing.
// It is the default implementation assigned to App.ReadPassword.
func ReadPassword(prompt string) ([]byte, error) {
	fmt.Print(prompt)
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return nil, err
	}
	return pass, nil
}

// RegisterCmd returns the cobra command for registering a new user.
func RegisterCmd(a *app.App) *cobra.Command {
	var username string
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			password, err := a.ReadPassword("Enter password: ")
			if err != nil {
				return err
			}
			defer crypto.Zero(password)

			masterPass, err := a.ReadPassword("Enter master password: ")
			if err != nil {
				return err
			}
			defer crypto.Zero(masterPass)

			salt := make([]byte, 16)
			if _, err := rand.Read(salt); err != nil {
				return err
			}

			dek, authPass, err := crypto.DeriveKeys(string(masterPass), salt)
			if err != nil {
				return err
			}
			defer crypto.Zero(dek)
			defer crypto.Zero(authPass)

			token, err := a.Client.Register(context.Background(), username, fmt.Sprintf("%x", authPass))
			if err != nil {
				return err
			}

			a.Storage.Token = token
			a.Storage.Salt = salt
			a.DEK = make([]byte, len(dek))
			copy(a.DEK, dek)
			return a.Storage.Save()
		},
	}
	cmd.Flags().StringVarP(&username, "user", "u", "", "username")
	cmd.MarkFlagRequired("user")
	return cmd
}

// LoginCmd returns the cobra command for authenticating an existing user.
func LoginCmd(a *app.App) *cobra.Command {
	var username string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			masterPass, err := a.ReadPassword("Enter master password: ")
			if err != nil {
				return err
			}
			defer crypto.Zero(masterPass)

			dek, authPass, err := crypto.DeriveKeys(string(masterPass), a.Storage.Salt)
			if err != nil {
				return err
			}
			defer crypto.Zero(dek)
			defer crypto.Zero(authPass)

			token, err := a.Client.Login(context.Background(), username, fmt.Sprintf("%x", authPass))
			if err != nil {
				return err
			}

			a.Storage.Token = token
			a.DEK = make([]byte, len(dek))
			copy(a.DEK, dek)
			a.Client.SetToken(token)
			return a.Storage.Save()
		},
	}
	cmd.Flags().StringVarP(&username, "user", "u", "", "username")
	cmd.MarkFlagRequired("user")
	return cmd
}
