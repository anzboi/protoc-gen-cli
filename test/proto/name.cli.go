package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func AccountsListAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "accountslistaccount",
		RunE: runAccountsListAccountCommand,
	}
	cmd.Flags().StringP("request", "d", "", "request object in json format")
	GetAccountListRequestFlags(cmd.Flags())
	return cmd
}

func runAccountsListAccountCommand(cmd *cobra.Command, args []string) error {
	// TODO: implement
	panic("unimplemented")
}

func AccountsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "accountslist",
	}
	cmd.AddCommand(AccountsListAccountCommand())
	return cmd
}

func GetAccountListRequestFlags(flags *pflag.FlagSet) {
	flags.String("filter", "", "")
}
