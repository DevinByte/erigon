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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/c2h5oh/datasize"
	lru "github.com/hashicorp/golang-lru/arc/v2"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/erigontech/mdbx-go/mdbx"
	"github.com/erigontech/secp256k1"

	chain2 "github.com/ledgerwatch/erigon-lib/chain"
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/datadir"
	"github.com/ledgerwatch/erigon-lib/common/dbg"
	"github.com/ledgerwatch/erigon-lib/config3"
	"github.com/ledgerwatch/erigon-lib/downloader"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/rawdbv3"
	"github.com/ledgerwatch/erigon-lib/log/v3"
	libstate "github.com/ledgerwatch/erigon-lib/state"
	"github.com/ledgerwatch/erigon-lib/wrap"

	"github.com/ledgerwatch/erigon/cl/clparams"
	"github.com/ledgerwatch/erigon/cmd/hack/tool/fromdb"
	"github.com/ledgerwatch/erigon/consensus"
	"github.com/ledgerwatch/erigon/core"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/rawdb/blockio"
	reset2 "github.com/ledgerwatch/erigon/core/rawdb/rawdbreset"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/core/vm"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/eth/ethconsensusconfig"
	"github.com/ledgerwatch/erigon/eth/integrity"
	"github.com/ledgerwatch/erigon/eth/stagedsync"
	"github.com/ledgerwatch/erigon/eth/stagedsync/stages"
	"github.com/ledgerwatch/erigon/ethdb/prune"
	"github.com/ledgerwatch/erigon/migrations"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/erigon/p2p"
	"github.com/ledgerwatch/erigon/p2p/sentry"
	"github.com/ledgerwatch/erigon/p2p/sentry/sentry_multi_client"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/polygon/bor"
	"github.com/ledgerwatch/erigon/polygon/heimdall"
	"github.com/ledgerwatch/erigon/turbo/builder"
	"github.com/ledgerwatch/erigon/turbo/debug"
	"github.com/ledgerwatch/erigon/turbo/logging"
	"github.com/ledgerwatch/erigon/turbo/services"
	"github.com/ledgerwatch/erigon/turbo/shards"
	"github.com/ledgerwatch/erigon/turbo/snapshotsync/freezeblocks"
	"github.com/ledgerwatch/erigon/turbo/snapshotsync/snap"
	stages2 "github.com/ledgerwatch/erigon/turbo/stages"
)

var cmdStageSnapshots = &cobra.Command{
	Use:   "stage_snapshots",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		if err := stageSnapshots(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdStageHeaders = &cobra.Command{
	Use:   "stage_headers",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		if err := stageHeaders(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdStageBorHeimdall = &cobra.Command{
	Use:   "stage_bor_heimdall",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		if err := stageBorHeimdall(db, cmd.Context(), unwindTypes, logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdStageBodies = &cobra.Command{
	Use:   "stage_bodies",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		if err := stageBodies(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdStageSenders = &cobra.Command{
	Use:   "stage_senders",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		if err := stageSenders(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdStageExec = &cobra.Command{
	Use:   "stage_exec",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		defer func(t time.Time) { logger.Info("total", "took", time.Since(t)) }(time.Now())

		if err := stageExec(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdStageCustomTrace = &cobra.Command{
	Use:   "stage_custom_trace",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		defer func(t time.Time) { logger.Info("total", "took", time.Since(t)) }(time.Now())

		if err := stageCustomTrace(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdStagePatriciaTrie = &cobra.Command{
	Use:   "rebuild_trie3_files",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		if err := stagePatriciaTrie(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdStageTxLookup = &cobra.Command{
	Use:   "stage_tx_lookup",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		if err := stageTxLookup(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdPrintStages = &cobra.Command{
	Use:   "print_stages",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Flags().Set(logging.LogConsoleVerbosityFlag.Name, "debug")
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), false, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		if err := printAllStages(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdPrintTableSizes = &cobra.Command{
	Use:   "print_table_sizes",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), false, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		allTablesCfg := db.AllTables()
		allTables := make([]string, 0, len(allTablesCfg))
		for table, cfg := range allTablesCfg {
			if cfg.IsDeprecated {
				continue
			}

			allTables = append(allTables, table)
		}

		var tableSizes []interface{}
		err = db.View(cmd.Context(), func(tx kv.Tx) error {
			tableSizes = stagedsync.CollectTableSizes(db, tx, allTables)
			return nil
		})
		if err != nil {
			logger.Error("error while collecting table sizes", "err", err)
			return
		}

		if len(tableSizes)%2 != 0 {
			logger.Error("table sizes len not even", "len", len(tableSizes))
			return
		}

		var sb strings.Builder
		sb.WriteString("Table")
		sb.WriteRune(',')
		sb.WriteString("Size")
		sb.WriteRune('\n')
		for i := 0; i < len(tableSizes)/2; i++ {
			sb.WriteString(tableSizes[i*2].(string))
			sb.WriteRune(',')
			sb.WriteString(tableSizes[i*2+1].(string))
			sb.WriteRune('\n')
		}

		if outputCsvFile == "" {
			logger.Info("table sizes", "csv", sb.String())
			return
		}

		f, err := os.Create(outputCsvFile)
		if err != nil {
			logger.Error("issue creating file", "file", outputCsvFile, "err", err)
			return
		}

		_, err = f.WriteString(sb.String())
		if err != nil {
			logger.Error("issue writing output to file", "file", outputCsvFile, "err", err)
			return
		}

		logger.Info("wrote table sizes to csv output file", "file", outputCsvFile)
	},
}

var cmdPrintMigrations = &cobra.Command{
	Use:   "print_migrations",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), false, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()
		if err := printAppliedMigrations(db, cmd.Context(), logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdRemoveMigration = &cobra.Command{
	Use:   "remove_migration",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), false, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()
		if err := removeMigration(db, cmd.Context()); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdRunMigrations = &cobra.Command{
	Use:   "run_migrations",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		migrations.EnableSqueezeCommitmentFiles = squeezeCommitmentFiles
		//non-accede and exclusive mode - to apply create new tables if need.
		cfg := dbCfg(kv.ChainDB, chaindata).Flags(func(u uint) uint { return u &^ mdbx.Accede }).Exclusive()
		db, err := openDB(cfg, true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()
		// Nothing to do, migrations will be applied automatically
	},
}

var cmdSetPrune = &cobra.Command{
	Use:   "force_set_prune",
	Short: "Override existing --prune flag value (if you know what you are doing)",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()
		if err := overrideStorageMode(db, logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

var cmdSetSnap = &cobra.Command{
	Use:   "force_set_snap",
	Short: "Override existing --snapshots flag value (if you know what you are doing)",
	Run: func(cmd *cobra.Command, args []string) {
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()
		sn, borSn, agg, _ := allSnapshots(cmd.Context(), db, logger)
		defer sn.Close()
		defer borSn.Close()
		defer agg.Close()

		cfg := sn.Cfg()
		flags := cmd.Flags()
		if flags.Lookup("snapshots") != nil {
			cfg.Enabled, err = flags.GetBool("snapshots")
			if err != nil {
				panic(err)
			}
		}

		if err := db.Update(context.Background(), func(tx kv.RwTx) error {
			return snap.ForceSetFlags(tx, cfg)
		}); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
			return
		}
	},
}

func init() {
	withConfig(cmdPrintStages)
	withDataDir(cmdPrintStages)
	withChain(cmdPrintStages)
	withHeimdall(cmdPrintStages)
	rootCmd.AddCommand(cmdPrintStages)

	withDataDir(cmdPrintTableSizes)
	withOutputCsvFile(cmdPrintTableSizes)
	rootCmd.AddCommand(cmdPrintTableSizes)

	withConfig(cmdStageSenders)
	withIntegrityChecks(cmdStageSenders)
	withReset(cmdStageSenders)
	withBlock(cmdStageSenders)
	withUnwind(cmdStageSenders)
	withDataDir(cmdStageSenders)
	withChain(cmdStageSenders)
	withHeimdall(cmdStageSenders)
	rootCmd.AddCommand(cmdStageSenders)

	withConfig(cmdStageSnapshots)
	withDataDir(cmdStageSnapshots)
	withChain(cmdStageSnapshots)
	withReset(cmdStageSnapshots)
	rootCmd.AddCommand(cmdStageSnapshots)

	withConfig(cmdStageHeaders)
	withIntegrityChecks(cmdStageHeaders)
	withDataDir(cmdStageHeaders)
	withUnwind(cmdStageHeaders)
	withReset(cmdStageHeaders)
	withChain(cmdStageHeaders)
	withHeimdall(cmdStageHeaders)
	rootCmd.AddCommand(cmdStageHeaders)

	withConfig(cmdStageBorHeimdall)
	withDataDir(cmdStageBorHeimdall)
	withReset(cmdStageBorHeimdall)
	withUnwind(cmdStageBorHeimdall)
	withChain(cmdStageBorHeimdall)
	withHeimdall(cmdStageBorHeimdall)
	rootCmd.AddCommand(cmdStageBorHeimdall)

	withConfig(cmdStageBodies)
	withDataDir(cmdStageBodies)
	withUnwind(cmdStageBodies)
	withChain(cmdStageBodies)
	withHeimdall(cmdStageBodies)
	rootCmd.AddCommand(cmdStageBodies)

	withConfig(cmdStageExec)
	withDataDir(cmdStageExec)
	withReset(cmdStageExec)
	withBlock(cmdStageExec)
	withUnwind(cmdStageExec)
	withNoCommit(cmdStageExec)
	withPruneTo(cmdStageExec)
	withBatchSize(cmdStageExec)
	withTxTrace(cmdStageExec)
	withChain(cmdStageExec)
	withHeimdall(cmdStageExec)
	withWorkers(cmdStageExec)
	rootCmd.AddCommand(cmdStageExec)

	withConfig(cmdStageCustomTrace)
	withDataDir(cmdStageCustomTrace)
	withReset(cmdStageCustomTrace)
	withBlock(cmdStageCustomTrace)
	withUnwind(cmdStageCustomTrace)
	withNoCommit(cmdStageCustomTrace)
	withPruneTo(cmdStageCustomTrace)
	withBatchSize(cmdStageCustomTrace)
	withTxTrace(cmdStageCustomTrace)
	withChain(cmdStageCustomTrace)
	withHeimdall(cmdStageCustomTrace)
	withWorkers(cmdStageCustomTrace)
	rootCmd.AddCommand(cmdStageCustomTrace)

	withConfig(cmdStagePatriciaTrie)
	withDataDir(cmdStagePatriciaTrie)
	withReset(cmdStagePatriciaTrie)
	withBlock(cmdStagePatriciaTrie)
	withUnwind(cmdStagePatriciaTrie)
	withPruneTo(cmdStagePatriciaTrie)
	withIntegrityChecks(cmdStagePatriciaTrie)
	withChain(cmdStagePatriciaTrie)
	withHeimdall(cmdStagePatriciaTrie)
	rootCmd.AddCommand(cmdStagePatriciaTrie)

	withConfig(cmdStageTxLookup)
	withReset(cmdStageTxLookup)
	withBlock(cmdStageTxLookup)
	withUnwind(cmdStageTxLookup)
	withDataDir(cmdStageTxLookup)
	withPruneTo(cmdStageTxLookup)
	withChain(cmdStageTxLookup)
	withHeimdall(cmdStageTxLookup)
	rootCmd.AddCommand(cmdStageTxLookup)

	withConfig(cmdPrintMigrations)
	withDataDir(cmdPrintMigrations)
	rootCmd.AddCommand(cmdPrintMigrations)

	withConfig(cmdRemoveMigration)
	withDataDir(cmdRemoveMigration)
	withMigration(cmdRemoveMigration)
	withChain(cmdRemoveMigration)
	withHeimdall(cmdRemoveMigration)
	rootCmd.AddCommand(cmdRemoveMigration)

	withConfig(cmdRunMigrations)
	withDataDir(cmdRunMigrations)
	withSqueezeCommitmentFiles(cmdRunMigrations)
	withChain(cmdRunMigrations)
	withHeimdall(cmdRunMigrations)
	rootCmd.AddCommand(cmdRunMigrations)

	withConfig(cmdSetSnap)
	withDataDir2(cmdSetSnap)
	withChain(cmdSetSnap)
	cmdSetSnap.Flags().Bool("snapshots", false, "")
	must(cmdSetSnap.MarkFlagRequired("snapshots"))
	rootCmd.AddCommand(cmdSetSnap)

	withConfig(cmdSetPrune)
	withDataDir(cmdSetPrune)
	withChain(cmdSetPrune)
	cmdSetPrune.Flags().StringVar(&pruneFlag, "prune", "hrtc", "")
	cmdSetPrune.Flags().Uint64Var(&pruneB, "prune.b.older", 0, "")
	cmdSetPrune.Flags().Uint64Var(&pruneH, "prune.h.older", 0, "")
	cmdSetPrune.Flags().Uint64Var(&pruneR, "prune.r.older", 0, "")
	cmdSetPrune.Flags().Uint64Var(&pruneT, "prune.t.older", 0, "")
	cmdSetPrune.Flags().Uint64Var(&pruneC, "prune.c.older", 0, "")
	cmdSetPrune.Flags().Uint64Var(&pruneBBefore, "prune.b.before", 0, "")
	cmdSetPrune.Flags().Uint64Var(&pruneHBefore, "prune.h.before", 0, "")
	cmdSetPrune.Flags().Uint64Var(&pruneRBefore, "prune.r.before", 0, "")
	cmdSetPrune.Flags().Uint64Var(&pruneTBefore, "prune.t.before", 0, "")
	cmdSetPrune.Flags().Uint64Var(&pruneCBefore, "prune.c.before", 0, "")
	cmdSetPrune.Flags().StringSliceVar(&experiments, "experiments", nil, "Storage mode to override database")
	cmdSetPrune.Flags().StringSliceVar(&unwindTypes, "unwind.types", nil, "Types to unwind for bor heimdall")
	rootCmd.AddCommand(cmdSetPrune)
}

func stageSnapshots(db kv.RwDB, ctx context.Context, logger log.Logger) error {
	sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
	defer sn.Close()
	defer borSn.Close()
	defer agg.Close()

	br, bw := blocksIO(db, logger)
	_, _, _, _, _ = newSync(ctx, db, nil /* miningConfig */, logger)
	chainConfig, _ := fromdb.ChainConfig(db), fromdb.PruneMode(db)

	return db.Update(ctx, func(tx kv.RwTx) error {
		if reset {
			if err := stages.SaveStageProgress(tx, stages.Snapshots, 0); err != nil {
				return fmt.Errorf("saving Snapshots progress failed: %w", err)
			}
		}
		dirs := datadir.New(datadirCli)
		if err := reset2.ResetBlocks(tx, db, agg, br, bw, dirs, *chainConfig, logger); err != nil {
			return fmt.Errorf("resetting blocks: %w", err)
		}
		ac := agg.BeginFilesRo()
		defer ac.Close()

		domains, err := libstate.NewSharedDomains(tx, logger)
		if err != nil {
			return err
		}
		defer domains.Close()
		//txnUm := domains.TxNum()
		blockNum := domains.BlockNum()

		// stagedsync.SpawnStageSnapshots(s, ctx, rwTx, logger)
		progress, err := stages.GetStageProgress(tx, stages.Snapshots)
		if err != nil {
			return fmt.Errorf("re-read Snapshots progress: %w", err)
		}

		if blockNum > progress {
			if err := stages.SaveStageProgress(tx, stages.Execution, blockNum); err != nil {
				return fmt.Errorf("saving Snapshots progress failed: %w", err)
			}
			progress, err = stages.GetStageProgress(tx, stages.Snapshots)
			if err != nil {
				return fmt.Errorf("re-read Snapshots progress: %w", err)
			}
		}
		logger.Info("Progress", "snapshots", progress)
		return nil
	})
}

func stageHeaders(db kv.RwDB, ctx context.Context, logger log.Logger) error {
	dirs := datadir.New(datadirCli)
	if err := datadir.ApplyMigrations(dirs); err != nil {
		return err
	}

	sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
	defer sn.Close()
	defer borSn.Close()
	defer agg.Close()
	br, bw := blocksIO(db, logger)
	_, _, _, _, _ = newSync(ctx, db, nil /* miningConfig */, logger)
	chainConfig, _ := fromdb.ChainConfig(db), fromdb.PruneMode(db)

	if integritySlow {
		if err := db.View(ctx, func(tx kv.Tx) error {
			log.Info("[integrity] no gaps in canonical headers")
			integrity.NoGapsInCanonicalHeaders(tx, ctx, br)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	if !(unwind > 0 || reset) {
		logger.Error("This command only works with --unwind or --reset options")
		return nil
	}

	return db.Update(ctx, func(tx kv.RwTx) error {
		if reset {
			if casted, ok := tx.(kv.CanWarmupDB); ok {
				if err := casted.WarmupDB(false); err != nil {
					return err
				}
			}

			if err := reset2.ResetBlocks(tx, db, agg, br, bw, dirs, *chainConfig, logger); err != nil {
				return err
			}
			return nil
		}

		progress, err := stages.GetStageProgress(tx, stages.Headers)
		if err != nil {
			return fmt.Errorf("read Bodies progress: %w", err)
		}
		var unwindTo uint64
		if unwind > progress {
			unwindTo = 1 // keep genesis
		} else {
			unwindTo = uint64(max(1, int(progress)-int(unwind)))
		}

		if err = stages.SaveStageProgress(tx, stages.Headers, unwindTo); err != nil {
			return fmt.Errorf("saving Headers progress failed: %w", err)
		}
		progress, err = stages.GetStageProgress(tx, stages.Headers)
		if err != nil {
			return fmt.Errorf("re-read Headers progress: %w", err)
		}
		{ // hard-unwind stage_body also
			if err := rawdb.TruncateBlocks(ctx, tx, progress+1); err != nil {
				return err
			}
			progressBodies, err := stages.GetStageProgress(tx, stages.Bodies)
			if err != nil {
				return fmt.Errorf("read Bodies progress: %w", err)
			}
			if progress < progressBodies {
				if err = stages.SaveStageProgress(tx, stages.Bodies, progress); err != nil {
					return fmt.Errorf("saving Bodies progress failed: %w", err)
				}
			}
		}
		// remove all canonical markers from this point
		if err = rawdb.TruncateCanonicalHash(tx, progress+1, false); err != nil {
			return err
		}
		if err = rawdb.TruncateTd(tx, progress+1); err != nil {
			return err
		}
		hash, err := br.CanonicalHash(ctx, tx, progress-1)
		if err != nil {
			return err
		}
		if err = rawdb.WriteHeadHeaderHash(tx, hash); err != nil {
			return err
		}

		logger.Info("Progress", "headers", progress)
		return nil
	})
}

func stageBorHeimdall(db kv.RwDB, ctx context.Context, unwindTypes []string, logger log.Logger) error {
	engine, _, sync, _, miningState := newSync(ctx, db, nil /* miningConfig */, logger)
	chainConfig := fromdb.ChainConfig(db)

	heimdallClient := engine.(*bor.Bor).HeimdallClient

	return db.Update(ctx, func(tx kv.RwTx) error {
		if reset {
			if err := reset2.ResetBorHeimdall(ctx, tx); err != nil {
				return err
			}
			return nil
		}
		if unwind > 0 {
			sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
			defer sn.Close()
			defer borSn.Close()
			defer agg.Close()

			stageState := stage(sync, tx, nil, stages.BorHeimdall)

			snapshotsMaxBlock := borSn.BlocksAvailable()
			if unwind <= snapshotsMaxBlock {
				return fmt.Errorf("cannot unwind past snapshots max block: %d", snapshotsMaxBlock)
			}

			if unwind > stageState.BlockNumber {
				return fmt.Errorf("cannot unwind to a point beyond stage: %d", stageState.BlockNumber)
			}

			unwindState := sync.NewUnwindState(stages.BorHeimdall, stageState.BlockNumber-unwind, stageState.BlockNumber, true, false)
			cfg := stagedsync.StageBorHeimdallCfg(db, nil, miningState, *chainConfig, nil, nil, nil, nil, nil, nil, nil, false, unwindTypes)
			if err := stagedsync.BorHeimdallUnwind(unwindState, ctx, stageState, tx, cfg); err != nil {
				return err
			}

			stageProgress, err := stages.GetStageProgress(tx, stages.BorHeimdall)
			if err != nil {
				return fmt.Errorf("re-read bor heimdall progress: %w", err)
			}

			logger.Info("progress", "bor heimdall", stageProgress)
			return nil
		}

		sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
		defer sn.Close()
		defer borSn.Close()
		defer agg.Close()
		blockReader, _ := blocksIO(db, logger)
		var (
			snapDb     kv.RwDB
			recents    *lru.ARCCache[libcommon.Hash, *bor.Snapshot]
			signatures *lru.ARCCache[libcommon.Hash, libcommon.Address]
		)
		if bor, ok := engine.(*bor.Bor); ok {
			snapDb = bor.DB
			recents = bor.Recents
			signatures = bor.Signatures
		}
		cfg := stagedsync.StageBorHeimdallCfg(db, snapDb, miningState, *chainConfig, heimdallClient, blockReader, nil, nil, nil, recents, signatures, false, unwindTypes)

		stageState := stage(sync, tx, nil, stages.BorHeimdall)
		if err := stagedsync.BorHeimdallForward(stageState, sync, ctx, tx, cfg, logger); err != nil {
			return err
		}

		stageProgress, err := stages.GetStageProgress(tx, stages.BorHeimdall)
		if err != nil {
			return fmt.Errorf("re-read bor heimdall progress: %w", err)
		}

		logger.Info("progress", "bor heimdall", stageProgress)
		return nil
	})
}

func stageBodies(db kv.RwDB, ctx context.Context, logger log.Logger) error {
	sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
	defer sn.Close()
	defer borSn.Close()
	defer agg.Close()
	chainConfig := fromdb.ChainConfig(db)
	_, _, sync, _, _ := newSync(ctx, db, nil /* miningConfig */, logger)
	br, bw := blocksIO(db, logger)

	if err := db.Update(ctx, func(tx kv.RwTx) error {
		s := stage(sync, tx, nil, stages.Bodies)

		if unwind > 0 {
			if unwind > s.BlockNumber {
				return fmt.Errorf("cannot unwind past 0")
			}

			u := sync.NewUnwindState(stages.Bodies, s.BlockNumber-unwind, s.BlockNumber, true, false)
			cfg := stagedsync.StageBodiesCfg(db, nil, nil, nil, nil, 0, *chainConfig, br, bw, nil)
			if err := stagedsync.UnwindBodiesStage(u, tx, cfg, ctx); err != nil {
				return err
			}

			progress, err := stages.GetStageProgress(tx, stages.Bodies)
			if err != nil {
				return fmt.Errorf("re-read Bodies progress: %w", err)
			}
			logger.Info("Progress", "bodies", progress)
			return nil
		}
		logger.Info("This command only works with --unwind option")
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func stageSenders(db kv.RwDB, ctx context.Context, logger log.Logger) error {
	tmpdir := datadir.New(datadirCli).Tmp
	chainConfig := fromdb.ChainConfig(db)
	sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
	defer sn.Close()
	defer borSn.Close()
	defer agg.Close()
	_, _, sync, _, _ := newSync(ctx, db, nil /* miningConfig */, logger)

	must(sync.SetCurrentStage(stages.Senders))

	br, _ := blocksIO(db, logger)
	if reset {
		return db.Update(ctx, func(tx kv.RwTx) error { return reset2.ResetSenders(ctx, db, tx) })
	}

	tx, err := db.BeginRw(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if integritySlow {
		secp256k1.ContextForThread(1)
		for i := block; ; i++ {
			if err := libcommon.Stopped(ctx.Done()); err != nil {
				return err
			}
			h, _ := br.HeaderByNumber(ctx, tx, i)
			if h == nil {
				break
			}
			withoutSenders, senders, err := br.BlockWithSenders(ctx, tx, h.Hash(), h.Number.Uint64())
			if err != nil {
				return err
			}
			withoutSenders.Body().SendersFromTxs() //remove senders info from txs
			txs := withoutSenders.Transactions()
			if txs.Len() != len(senders) {
				logger.Error("not equal amount of senders", "block", i, "db", len(senders), "expect", txs.Len())
				return nil
			}
			if txs.Len() == 0 {
				continue
			}
			signer := types.MakeSigner(chainConfig, i, h.Time)
			for j := 0; j < txs.Len(); j++ {
				from, err := signer.Sender(txs[j])
				if err != nil {
					return err
				}
				if !bytes.Equal(from[:], senders[j][:]) {
					logger.Error("wrong sender", "block", i, "tx", j, "db", fmt.Sprintf("%x", senders[j]), "expect", fmt.Sprintf("%x", from))
				}
			}
			if i%10 == 0 {
				logger.Info("checked", "block", i)
			}
		}
		return nil
	}

	s := stage(sync, tx, nil, stages.Senders)
	logger.Info("Stage", "name", s.ID, "progress", s.BlockNumber)

	pm, err := prune.Get(tx)
	if err != nil {
		return err
	}

	cfg := stagedsync.StageSendersCfg(db, chainConfig, sync.Cfg(), false, tmpdir, pm, br, nil, nil)
	if unwind > 0 {
		u := sync.NewUnwindState(stages.Senders, s.BlockNumber-unwind, s.BlockNumber, true, false)
		if err = stagedsync.UnwindSendersStage(u, tx, cfg, ctx); err != nil {
			return err
		}
	} else if pruneTo > 0 {
		p, err := sync.PruneStageState(stages.Senders, s.BlockNumber, tx, db, true)
		if err != nil {
			return err
		}
		if err = stagedsync.PruneSendersStage(p, tx, cfg, ctx); err != nil {
			return err
		}
		return nil
	} else {
		if err = stagedsync.SpawnRecoverSendersStage(cfg, s, sync, tx, block, ctx, logger); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func stageExec(db kv.RwDB, ctx context.Context, logger log.Logger) error {
	dirs := datadir.New(datadirCli)
	if err := datadir.ApplyMigrations(dirs); err != nil {
		return err
	}

	engine, vmConfig, sync, _, _ := newSync(ctx, db, nil /* miningConfig */, logger)
	must(sync.SetCurrentStage(stages.Execution))
	sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
	defer sn.Close()
	defer borSn.Close()
	defer agg.Close()
	if warmup {
		return reset2.WarmupExec(ctx, db)
	}
	if reset {
		if err := reset2.ResetExec(ctx, db, chain, "", logger); err != nil {
			return err
		}
		return nil
	}

	if txtrace {
		// Activate tracing and writing into json files for each transaction
		vmConfig.Tracer = nil
		vmConfig.Debug = true
	}

	var batchSize datasize.ByteSize
	must(batchSize.UnmarshalText([]byte(batchSizeStr)))

	s := stage(sync, nil, db, stages.Execution)

	logger.Info("Stage", "name", s.ID, "progress", s.BlockNumber)
	chainConfig, pm := fromdb.ChainConfig(db), fromdb.PruneMode(db)
	if pruneTo > 0 {
		pm.History = prune.Distance(s.BlockNumber - pruneTo)
		pm.Receipts = prune.Distance(s.BlockNumber - pruneTo)
		pm.CallTraces = prune.Distance(s.BlockNumber - pruneTo)
		pm.TxIndex = prune.Distance(s.BlockNumber - pruneTo)
	}

	syncCfg := ethconfig.Defaults.Sync
	syncCfg.ExecWorkerCount = int(workers)
	syncCfg.ReconWorkerCount = int(reconWorkers)

	genesis := core.GenesisBlockByChainName(chain)
	br, _ := blocksIO(db, logger)
	cfg := stagedsync.StageExecuteBlocksCfg(db, pm, batchSize, chainConfig, engine, vmConfig, nil,
		/*stateStream=*/ false,
		/*badBlockHalt=*/ true, dirs, br, nil, genesis, syncCfg, agg, nil)

	if unwind > 0 {
		if err := db.View(ctx, func(tx kv.Tx) error {
			minUnwindableBlockNum, _, err := tx.(libstate.HasAggTx).AggTx().(*libstate.AggregatorRoTx).CanUnwindBeforeBlockNum(s.BlockNumber-unwind, tx)
			if err != nil {
				return err
			}
			unwind = s.BlockNumber - minUnwindableBlockNum
			return nil
		}); err != nil {
			return err
		}
	}

	var tx kv.RwTx //nil - means lower-level code (each stage) will manage transactions
	if noCommit {
		var err error
		tx, err = db.BeginRw(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}
	txc := wrap.TxContainer{Tx: tx}

	if unwind > 0 {
		u := sync.NewUnwindState(stages.Execution, s.BlockNumber-unwind, s.BlockNumber, true, false)
		err := stagedsync.UnwindExecutionStage(u, s, txc, ctx, cfg, logger)
		if err != nil {
			return err
		}
		return nil
	}

	if pruneTo > 0 {
		p, err := sync.PruneStageState(stages.Execution, s.BlockNumber, tx, db, true)
		if err != nil {
			return err
		}
		err = stagedsync.PruneExecutionStage(p, tx, cfg, ctx)
		if err != nil {
			return err
		}
		return nil
	}

	err := stagedsync.SpawnExecuteBlocksStage(s, sync, txc, block, ctx, cfg, logger)
	if err != nil {
		return err
	}

	return nil
}

func stageCustomTrace(db kv.RwDB, ctx context.Context, logger log.Logger) error {
	dirs := datadir.New(datadirCli)
	if err := datadir.ApplyMigrations(dirs); err != nil {
		return err
	}

	engine, vmConfig, sync, _, _ := newSync(ctx, db, nil /* miningConfig */, logger)
	must(sync.SetCurrentStage(stages.Execution))
	sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
	defer sn.Close()
	defer borSn.Close()
	defer agg.Close()
	if warmup {
		panic("not implemented")
		//return reset2.WarmupExec(ctx, db)
	}
	if reset {
		if err := reset2.Reset(ctx, db, stages.CustomTrace); err != nil {
			return err
		}
		return nil
	}

	if txtrace {
		// Activate tracing and writing into json files for each transaction
		vmConfig.Tracer = nil
		vmConfig.Debug = true
	}

	var batchSize datasize.ByteSize
	must(batchSize.UnmarshalText([]byte(batchSizeStr)))

	s := stage(sync, nil, db, stages.CustomTrace)

	logger.Info("Stage", "name", s.ID, "progress", s.BlockNumber)
	chainConfig, pm := fromdb.ChainConfig(db), fromdb.PruneMode(db)
	if pruneTo > 0 {
		pm.History = prune.Distance(s.BlockNumber - pruneTo)
		pm.Receipts = prune.Distance(s.BlockNumber - pruneTo)
		pm.CallTraces = prune.Distance(s.BlockNumber - pruneTo)
		pm.TxIndex = prune.Distance(s.BlockNumber - pruneTo)
	}

	syncCfg := ethconfig.Defaults.Sync
	syncCfg.ExecWorkerCount = int(workers)
	syncCfg.ReconWorkerCount = int(reconWorkers)

	genesis := core.GenesisBlockByChainName(chain)
	br, _ := blocksIO(db, logger)
	cfg := stagedsync.StageCustomTraceCfg(db, pm, dirs, br, chainConfig, engine, genesis, &syncCfg)

	if unwind > 0 {
		if err := db.View(ctx, func(tx kv.Tx) error {
			blockNumWithCommitment, ok, err := tx.(libstate.HasAggTx).AggTx().(*libstate.AggregatorRoTx).CanUnwindBeforeBlockNum(s.BlockNumber-unwind, tx)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("too deep unwind requested: %d, minimum allowed: %d", s.BlockNumber-unwind, blockNumWithCommitment)
			}
			unwind = s.BlockNumber - blockNumWithCommitment
			return nil
		}); err != nil {
			return err
		}
	}

	var tx kv.RwTx //nil - means lower-level code (each stage) will manage transactions
	if noCommit {
		var err error
		tx, err = db.BeginRw(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}
	txc := wrap.TxContainer{Tx: tx}

	if unwind > 0 {
		u := sync.NewUnwindState(stages.CustomTrace, s.BlockNumber-unwind, s.BlockNumber, true, false)
		err := stagedsync.UnwindCustomTrace(u, s, txc, cfg, ctx, logger)
		if err != nil {
			return err
		}
		return nil
	}

	if pruneTo > 0 {
		p, err := sync.PruneStageState(stages.CustomTrace, s.BlockNumber, tx, db, true)
		if err != nil {
			return err
		}
		err = stagedsync.PruneCustomTrace(p, tx, cfg, ctx, logger)
		if err != nil {
			return err
		}
		return nil
	}

	err := stagedsync.SpawnCustomTrace(s, txc, cfg, ctx, block, logger)
	if err != nil {
		return err
	}

	return nil
}

func stagePatriciaTrie(db kv.RwDB, ctx context.Context, logger log.Logger) error {
	dirs, pm := datadir.New(datadirCli), fromdb.PruneMode(db)
	_ = pm
	sn, _, agg, _ := allSnapshots(ctx, db, logger)
	defer sn.Close()
	defer agg.Close()
	_, _, _, _, _ = newSync(ctx, db, nil /* miningConfig */, logger)

	if warmup {
		return reset2.Warmup(ctx, db, log.LvlInfo, stages.Execution)
	}
	if reset {
		return reset2.Reset(ctx, db, stages.Execution)
	}
	tx, err := db.BeginRw(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	br, _ := blocksIO(db, logger)
	historyV3 := true
	cfg := stagedsync.StageTrieCfg(db, true /* checkRoot */, true /* saveHashesToDb */, false /* badBlockHalt */, dirs.Tmp, br, nil /* hd */, historyV3, agg)

	if _, err := stagedsync.RebuildPatriciaTrieBasedOnFiles(tx, cfg, ctx, logger); err != nil {
		return err
	}
	return tx.Commit()
}

func stageTxLookup(db kv.RwDB, ctx context.Context, logger log.Logger) error {
	dirs, pm := datadir.New(datadirCli), fromdb.PruneMode(db)
	_, _, sync, _, _ := newSync(ctx, db, nil /* miningConfig */, logger)
	chainConfig := fromdb.ChainConfig(db)
	must(sync.SetCurrentStage(stages.TxLookup))
	sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
	defer sn.Close()
	defer borSn.Close()
	defer agg.Close()

	if reset {
		return db.Update(ctx, func(tx kv.RwTx) error { return reset2.ResetTxLookup(tx) })
	}
	tx, err := db.BeginRw(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	s := stage(sync, tx, nil, stages.TxLookup)
	if pruneTo > 0 {
		pm.History = prune.Distance(s.BlockNumber - pruneTo)
		pm.Receipts = prune.Distance(s.BlockNumber - pruneTo)
		pm.CallTraces = prune.Distance(s.BlockNumber - pruneTo)
		pm.TxIndex = prune.Distance(s.BlockNumber - pruneTo)
	}
	logger.Info("Stage", "name", s.ID, "progress", s.BlockNumber)

	br, _ := blocksIO(db, logger)
	cfg := stagedsync.StageTxLookupCfg(db, pm, dirs.Tmp, chainConfig.Bor, br)
	if unwind > 0 {
		u := sync.NewUnwindState(stages.TxLookup, s.BlockNumber-unwind, s.BlockNumber, true, false)
		err = stagedsync.UnwindTxLookup(u, s, tx, cfg, ctx, logger)
		if err != nil {
			return err
		}
	} else if pruneTo > 0 {
		p, err := sync.PruneStageState(stages.TxLookup, s.BlockNumber, tx, nil, true)
		if err != nil {
			return err
		}
		err = stagedsync.PruneTxLookup(p, tx, cfg, ctx, logger)
		if err != nil {
			return err
		}
	} else {
		err = stagedsync.SpawnTxLookup(s, tx, block, cfg, ctx, logger)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func printAllStages(db kv.RoDB, ctx context.Context, logger log.Logger) error {
	sn, borSn, agg, _ := allSnapshots(ctx, db, logger)
	defer sn.Close()
	defer borSn.Close()
	defer agg.Close()
	return db.View(ctx, func(tx kv.Tx) error { return printStages(tx, sn, borSn, agg) })
}

func printAppliedMigrations(db kv.RwDB, ctx context.Context, logger log.Logger) error {
	return db.View(ctx, func(tx kv.Tx) error {
		applied, err := migrations.AppliedMigrations(tx, false /* withPayload */)
		if err != nil {
			return err
		}
		var appliedStrs = make([]string, len(applied))
		i := 0
		for k := range applied {
			appliedStrs[i] = k
			i++
		}
		slices.Sort(appliedStrs)
		logger.Info("Applied", "migrations", strings.Join(appliedStrs, " "))
		return nil
	})
}

func removeMigration(db kv.RwDB, ctx context.Context) error {
	return db.Update(ctx, func(tx kv.RwTx) error {
		return tx.Delete(kv.Migrations, []byte(migration))
	})
}

var openSnapshotOnce sync.Once
var _allSnapshotsSingleton *freezeblocks.RoSnapshots
var _allBorSnapshotsSingleton *freezeblocks.BorRoSnapshots
var _allCaplinSnapshotsSingleton *freezeblocks.CaplinSnapshots
var _aggSingleton *libstate.Aggregator

func allSnapshots(ctx context.Context, db kv.RoDB, logger log.Logger) (*freezeblocks.RoSnapshots, *freezeblocks.BorRoSnapshots, *libstate.Aggregator, *freezeblocks.CaplinSnapshots) {
	openSnapshotOnce.Do(func() {
		var useSnapshots bool
		_ = db.View(context.Background(), func(tx kv.Tx) error {
			useSnapshots, _ = snap.Enabled(tx)
			return nil
		})
		dirs := datadir.New(datadirCli)

		//useSnapshots = true
		snapCfg := ethconfig.NewSnapCfg(useSnapshots, true, true, true)

		_allSnapshotsSingleton = freezeblocks.NewRoSnapshots(snapCfg, dirs.Snap, 0, logger)
		_allBorSnapshotsSingleton = freezeblocks.NewBorRoSnapshots(snapCfg, dirs.Snap, 0, logger)
		var err error
		cr := rawdb.NewCanonicalReader()
		_aggSingleton, err = libstate.NewAggregator(ctx, dirs, config3.HistoryV3AggregationStep, db, cr, logger)
		if err != nil {
			panic(err)
		}

		_aggSingleton.SetProduceMod(snapCfg.ProduceE3)

		if useSnapshots {
			g := &errgroup.Group{}
			g.Go(func() error {
				_allSnapshotsSingleton.OptimisticalyReopenFolder()
				return nil
			})
			g.Go(func() error {
				_allBorSnapshotsSingleton.OptimisticalyReopenFolder()
				return nil
			})
			g.Go(func() error { return _aggSingleton.OpenFolder() })
			g.Go(func() error {
				chainConfig := fromdb.ChainConfig(db)
				var beaconConfig *clparams.BeaconChainConfig
				_, beaconConfig, _, err = clparams.GetConfigsByNetworkName(chainConfig.ChainName)
				if err == nil {
					_allCaplinSnapshotsSingleton = freezeblocks.NewCaplinSnapshots(snapCfg, beaconConfig, dirs, logger)
					if err = _allCaplinSnapshotsSingleton.ReopenFolder(); err != nil {
						return err
					}
					_allCaplinSnapshotsSingleton.LogStat("caplin")
				}
				return nil
			})

			g.Go(func() error {
				ls, er := os.Stat(filepath.Join(dirs.Snap, downloader.ProhibitNewDownloadsFileName))
				mtime := time.Time{}
				if er == nil {
					mtime = ls.ModTime()
				}
				logger.Info("[downloads]", "locked", er == nil, "at", mtime.Format("02 Jan 06 15:04 2006"))
				return nil
			})
			err := g.Wait()
			if err != nil {
				panic(err)
			}

			_allSnapshotsSingleton.LogStat("blocks")
			_allBorSnapshotsSingleton.LogStat("bor")
			_ = db.View(context.Background(), func(tx kv.Tx) error {
				ac := _aggSingleton.BeginFilesRo()
				defer ac.Close()
				ac.LogStats(tx, func(endTxNumMinimax uint64) (uint64, error) {
					_, histBlockNumProgress, err := rawdbv3.TxNums.FindBlockNum(tx, endTxNumMinimax)
					return histBlockNumProgress, err
				})
				return nil
			})
		}
	})
	return _allSnapshotsSingleton, _allBorSnapshotsSingleton, _aggSingleton, _allCaplinSnapshotsSingleton
}

var openBlockReaderOnce sync.Once
var _blockReaderSingleton services.FullBlockReader
var _blockWriterSingleton *blockio.BlockWriter

func blocksIO(db kv.RoDB, logger log.Logger) (services.FullBlockReader, *blockio.BlockWriter) {
	openBlockReaderOnce.Do(func() {
		sn, borSn, _, _ := allSnapshots(context.Background(), db, logger)
		_blockReaderSingleton = freezeblocks.NewBlockReader(sn, borSn)
		_blockWriterSingleton = blockio.NewBlockWriter()
	})
	return _blockReaderSingleton, _blockWriterSingleton
}

const blockBufferSize = 128

func newSync(ctx context.Context, db kv.RwDB, miningConfig *params.MiningConfig, logger log.Logger) (consensus.Engine, *vm.Config, *stagedsync.Sync, *stagedsync.Sync, stagedsync.MiningState) {
	dirs, pm := datadir.New(datadirCli), fromdb.PruneMode(db)

	vmConfig := &vm.Config{}

	events := shards.NewEvents()

	genesis := core.GenesisBlockByChainName(chain)
	chainConfig, genesisBlock, genesisErr := core.CommitGenesisBlock(db, genesis, "", logger)
	if _, ok := genesisErr.(*chain2.ConfigCompatError); genesisErr != nil && !ok {
		panic(genesisErr)
	}
	//logger.Info("Initialised chain configuration", "config", chainConfig)

	var batchSize datasize.ByteSize
	must(batchSize.UnmarshalText([]byte(batchSizeStr)))

	cfg := ethconfig.Defaults
	cfg.Prune = pm
	cfg.BatchSize = batchSize
	cfg.DeprecatedTxPool.Disable = true
	cfg.Genesis = genesis
	if miningConfig != nil {
		cfg.Miner = *miningConfig
	}
	cfg.Dirs = datadir.New(datadirCli)
	allSn, _, agg, _ := allSnapshots(ctx, db, logger)
	cfg.Snapshot = allSn.Cfg()

	blockReader, blockWriter := blocksIO(db, logger)
	engine, heimdallClient := initConsensusEngine(ctx, chainConfig, cfg.Dirs.DataDir, db, blockReader, logger)

	statusDataProvider := sentry.NewStatusDataProvider(
		db,
		chainConfig,
		genesisBlock,
		chainConfig.ChainID.Uint64(),
		logger,
	)

	maxBlockBroadcastPeers := func(header *types.Header) uint { return 0 }

	sentryControlServer, err := sentry_multi_client.NewMultiClient(
		db,
		chainConfig,
		engine,
		nil,
		ethconfig.Defaults.Sync,
		blockReader,
		blockBufferSize,
		statusDataProvider,
		false,
		maxBlockBroadcastPeers,
		false, /* disableBlockDownload */
		logger,
	)
	if err != nil {
		panic(err)
	}

	blockSnapBuildSema := semaphore.NewWeighted(int64(dbg.BuildSnapshotAllowance))
	agg.SetSnapshotBuildSema(blockSnapBuildSema)

	notifications := &shards.Notifications{}
	blockRetire := freezeblocks.NewBlockRetire(1, dirs, blockReader, blockWriter, db, chainConfig, notifications.Events, blockSnapBuildSema, logger)

	var (
		snapDb     kv.RwDB
		recents    *lru.ARCCache[libcommon.Hash, *bor.Snapshot]
		signatures *lru.ARCCache[libcommon.Hash, libcommon.Address]
	)
	if bor, ok := engine.(*bor.Bor); ok {
		snapDb = bor.DB
		recents = bor.Recents
		signatures = bor.Signatures
	}
	stages := stages2.NewDefaultStages(context.Background(), db, snapDb, p2p.Config{}, &cfg, sentryControlServer, notifications, nil, blockReader, blockRetire, agg, nil, nil,
		heimdallClient, recents, signatures, logger)
	sync := stagedsync.New(cfg.Sync, stages, stagedsync.DefaultUnwindOrder, stagedsync.DefaultPruneOrder, logger)

	miner := stagedsync.NewMiningState(&cfg.Miner)
	miningCancel := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(miningCancel)
	}()

	miningSync := stagedsync.New(
		cfg.Sync,
		stagedsync.MiningStages(ctx,
			stagedsync.StageMiningCreateBlockCfg(db, miner, *chainConfig, engine, nil, nil, dirs.Tmp, blockReader),
			stagedsync.StageBorHeimdallCfg(db, snapDb, miner, *chainConfig, heimdallClient, blockReader, nil, nil, nil, recents, signatures, false, unwindTypes),
			stagedsync.StageExecuteBlocksCfg(
				db,
				cfg.Prune,
				cfg.BatchSize,
				sentryControlServer.ChainConfig,
				sentryControlServer.Engine,
				&vm.Config{},
				notifications.Accumulator,
				cfg.StateStream,
				/*stateStream=*/ false,
				dirs,
				blockReader,
				sentryControlServer.Hd,
				cfg.Genesis,
				cfg.Sync,
				agg,
				nil,
			),
			stagedsync.StageSendersCfg(db, sentryControlServer.ChainConfig, cfg.Sync, false, dirs.Tmp, cfg.Prune, blockReader, sentryControlServer.Hd, nil),
			stagedsync.StageMiningExecCfg(db, miner, events, *chainConfig, engine, &vm.Config{}, dirs.Tmp, nil, 0, nil, nil, blockReader),
			stagedsync.StageMiningFinishCfg(db, *chainConfig, engine, miner, miningCancel, blockReader, builder.NewLatestBlockBuiltStore()),
		),
		stagedsync.MiningUnwindOrder,
		stagedsync.MiningPruneOrder,
		logger,
	)

	return engine, vmConfig, sync, miningSync, miner
}

func progress(tx kv.Getter, stage stages.SyncStage) uint64 {
	res, err := stages.GetStageProgress(tx, stage)
	if err != nil {
		panic(err)
	}
	return res
}

func stage(st *stagedsync.Sync, tx kv.Tx, db kv.RoDB, stage stages.SyncStage) *stagedsync.StageState {
	res, err := st.StageState(stage, tx, db, true, false)
	if err != nil {
		panic(err)
	}
	return res
}

func overrideStorageMode(db kv.RwDB, logger log.Logger) error {
	chainConfig := fromdb.ChainConfig(db)
	pm, err := prune.FromCli(chainConfig.ChainID.Uint64(), pruneFlag, pruneB, pruneH, pruneR, pruneT, pruneC,
		pruneHBefore, pruneRBefore, pruneTBefore, pruneCBefore, pruneBBefore, experiments)
	if err != nil {
		return err
	}
	return db.Update(context.Background(), func(tx kv.RwTx) error {
		if err = prune.Override(tx, pm); err != nil {
			return err
		}
		pm, err = prune.Get(tx)
		if err != nil {
			return err
		}
		logger.Info("Storage mode in DB", "mode", pm.String())
		return nil
	})
}

func initConsensusEngine(ctx context.Context, cc *chain2.Config, dir string, db kv.RwDB, blockReader services.FullBlockReader, logger log.Logger) (engine consensus.Engine, heimdallClient heimdall.HeimdallClient) {
	config := ethconfig.Defaults

	var consensusConfig interface{}
	if cc.Clique != nil {
		consensusConfig = params.CliqueSnapshot
	} else if cc.Aura != nil {
		consensusConfig = &config.Aura
	} else if cc.Bor != nil {
		consensusConfig = cc.Bor
		config.HeimdallURL = HeimdallURL
		if !config.WithoutHeimdall {
			heimdallClient = heimdall.NewHeimdallClient(config.HeimdallURL, logger)
		}
	} else {
		consensusConfig = &config.Ethash
	}
	return ethconsensusconfig.CreateConsensusEngine(ctx, &nodecfg.Config{Dirs: datadir.New(dir)}, cc, consensusConfig, config.Miner.Notify, config.Miner.Noverify,
		heimdallClient, config.WithoutHeimdall, blockReader, db.ReadOnly(), logger, nil), heimdallClient
}
