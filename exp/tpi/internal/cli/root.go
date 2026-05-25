package cli

import (
	"github.com/spf13/cobra"
)

func RatesCmd() *cobra.Command {
	return ratesCmd()
}

func UpdateCmd() *cobra.Command {
	return updateCmd()
}
