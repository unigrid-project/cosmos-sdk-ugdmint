package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/unigrid-project/cosmos-sdk-ugdmint/x/ugdmint/types"
	"github.com/spf13/cobra"
)

// cmdQuerySubsidyHalvingInterval implements a command to return the current minting
// subsidy halving interval value.
func cmdQuerySubsidyHalvingInterval() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subsidy-halving-interval",
		Short: "Query the current minting subsidy halving interval",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QuerySubsidyHalvingIntervalRequest{}
			res, err := queryClient.SubsidyHalvingInterval(cmd.Context(), params)

			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.SubsidyHalvingInterval))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
