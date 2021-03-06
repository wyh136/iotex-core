// Copyright (c) 2018 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotexproject/iotex-core/logger"
)

var limit int

// transfersCmd represents the transfers command
var transfersCmd = &cobra.Command{
	Use:   "transfers [addr]",
	Short: "Returns the transfers associated with a given address",
	Long:  `Returns the transfers associated with a given address`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(transfers(args))
	},
}

func transfers(args []string) string {
	client, _ := getClientAndCfg()
	transfers, err := client.GetTransfersByAddress(args[0], 0, int64(limit))
	if err != nil {
		logger.Error().Err(err).Msgf("cannot get transfers for address %s", args[0])
		return ""
	}
	var res string
	for _, t := range transfers {
		res += fmt.Sprintf("%+v\n", t)
	}
	return res[:len(res)-1] // get rid of the extra \n at the end
}

func init() {
	rootCmd.AddCommand(transfersCmd)
	transfersCmd.PersistentFlags().IntVarP(&limit, "limit", "l", 1000000000, "max transfers to display")
}
