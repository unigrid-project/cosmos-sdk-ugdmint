package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/unigrid-project/cosmos-ugdmint/x/ugdmint/types"
)

func cmdQueryMints() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mints",
		Short: "Query the minting records",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllMintRecordsRequest{}
			res, err := queryClient.AllMintRecords(cmd.Context(), params)

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
