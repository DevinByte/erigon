// Copyright 2024 The Erigon Authors
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package commands

import (
	"github.com/ledgerwatch/erigon/cmd/state/verify"
	"github.com/ledgerwatch/erigon/turbo/debug"
	"github.com/spf13/cobra"
)

func init() {
	withDataDir(verifyTxLookupCmd)
	rootCmd.AddCommand(verifyTxLookupCmd)
}

var verifyTxLookupCmd = &cobra.Command{
	Use:   "verifyTxLookup",
	Short: "Generate txn lookup index",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := debug.SetupCobra(cmd, "verify_txlookup")
		return verify.ValidateTxLookups(chaindata, logger)
	},
}
