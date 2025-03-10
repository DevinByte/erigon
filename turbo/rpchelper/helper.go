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

package rpchelper

import (
	"context"
	"errors"
	"fmt"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/kvcache"
	"github.com/ledgerwatch/erigon-lib/kv/rawdbv3"
	"github.com/ledgerwatch/erigon-lib/wrap"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/state"
	"github.com/ledgerwatch/erigon/eth/stagedsync/stages"
	borfinality "github.com/ledgerwatch/erigon/polygon/bor/finality"
	"github.com/ledgerwatch/erigon/polygon/bor/finality/whitelist"
	"github.com/ledgerwatch/erigon/rpc"
)

// unable to decode supplied params, or an invalid number of parameters
type nonCanonocalHashError struct{ hash libcommon.Hash }

func (e nonCanonocalHashError) ErrorCode() int { return -32603 }

func (e nonCanonocalHashError) Error() string {
	return fmt.Sprintf("hash %x is not currently canonical", e.hash)
}

func GetBlockNumber(blockNrOrHash rpc.BlockNumberOrHash, tx kv.Tx, filters *Filters) (uint64, libcommon.Hash, bool, error) {
	return _GetBlockNumber(blockNrOrHash.RequireCanonical, blockNrOrHash, tx, filters)
}

func GetCanonicalBlockNumber(blockNrOrHash rpc.BlockNumberOrHash, tx kv.Tx, filters *Filters) (uint64, libcommon.Hash, bool, error) {
	return _GetBlockNumber(true, blockNrOrHash, tx, filters)
}

func _GetBlockNumber(requireCanonical bool, blockNrOrHash rpc.BlockNumberOrHash, tx kv.Tx, filters *Filters) (blockNumber uint64, hash libcommon.Hash, latest bool, err error) {
	// Due to changed semantics of `lastest` block in RPC request, it is now distinct
	// from the block number corresponding to the plain state
	var plainStateBlockNumber uint64
	if plainStateBlockNumber, err = stages.GetStageProgress(tx, stages.Execution); err != nil {
		return 0, libcommon.Hash{}, false, fmt.Errorf("getting plain state block number: %w", err)
	}
	var ok bool
	hash, ok = blockNrOrHash.Hash()
	if !ok {
		number := *blockNrOrHash.BlockNumber
		switch number {
		case rpc.LatestBlockNumber:
			if blockNumber, err = GetLatestBlockNumber(tx); err != nil {
				return 0, libcommon.Hash{}, false, err
			}
		case rpc.EarliestBlockNumber:
			blockNumber = 0
		case rpc.FinalizedBlockNumber:
			if whitelist.GetWhitelistingService() != nil {
				num := borfinality.GetFinalizedBlockNumber(tx)
				if num == 0 {
					// nolint
					return 0, libcommon.Hash{}, false, errors.New("No finalized block")
				}

				blockNum := borfinality.CurrentFinalizedBlock(tx, num).NumberU64()
				blockHash := rawdb.ReadHeaderByNumber(tx, blockNum).Hash()
				return blockNum, blockHash, false, nil
			}
			blockNumber, err = GetFinalizedBlockNumber(tx)
			if err != nil {
				return 0, libcommon.Hash{}, false, err
			}
		case rpc.SafeBlockNumber:
			blockNumber, err = GetSafeBlockNumber(tx)
			if err != nil {
				return 0, libcommon.Hash{}, false, err
			}
		case rpc.PendingBlockNumber:
			pendingBlock := filters.LastPendingBlock()
			if pendingBlock == nil {
				blockNumber = plainStateBlockNumber
			} else {
				return pendingBlock.NumberU64(), pendingBlock.Hash(), false, nil
			}
		case rpc.LatestExecutedBlockNumber:
			blockNumber = plainStateBlockNumber
		default:
			blockNumber = uint64(number.Int64())
		}
		hash, err = rawdb.ReadCanonicalHash(tx, blockNumber)
		if err != nil {
			return 0, libcommon.Hash{}, false, err
		}
	} else {
		number := rawdb.ReadHeaderNumber(tx, hash)
		if number == nil {
			return 0, libcommon.Hash{}, false, fmt.Errorf("block %x not found", hash)
		}
		blockNumber = *number

		ch, err := rawdb.ReadCanonicalHash(tx, blockNumber)
		if err != nil {
			return 0, libcommon.Hash{}, false, err
		}
		if requireCanonical && ch != hash {
			return 0, libcommon.Hash{}, false, nonCanonocalHashError{hash}
		}
	}
	return blockNumber, hash, blockNumber == plainStateBlockNumber, nil
}

func CreateStateReader(ctx context.Context, tx kv.Tx, blockNrOrHash rpc.BlockNumberOrHash, txnIndex int, filters *Filters, stateCache kvcache.Cache, chainName string) (state.StateReader, error) {
	blockNumber, _, latest, err := _GetBlockNumber(true, blockNrOrHash, tx, filters)
	if err != nil {
		return nil, err
	}
	return CreateStateReaderFromBlockNumber(ctx, tx, blockNumber, latest, txnIndex, stateCache, chainName)
}

func CreateStateReaderFromBlockNumber(ctx context.Context, tx kv.Tx, blockNumber uint64, latest bool, txnIndex int, stateCache kvcache.Cache, chainName string) (state.StateReader, error) {
	if latest {
		cacheView, err := stateCache.View(ctx, tx)
		if err != nil {
			return nil, err
		}
		return CreateLatestCachedStateReader(cacheView, tx), nil
	}
	return CreateHistoryStateReader(tx, blockNumber+1, txnIndex, chainName)
}

func CreateHistoryStateReader(tx kv.Tx, blockNumber uint64, txnIndex int, chainName string) (state.StateReader, error) {
	r := state.NewHistoryReaderV3()
	r.SetTx(tx)
	//r.SetTrace(true)
	minTxNum, err := rawdbv3.TxNums.Min(tx, blockNumber)
	if err != nil {
		return nil, err
	}
	r.SetTxNum(uint64(int(minTxNum) + txnIndex + /* 1 system txNum in beginning of block */ 1))
	return r, nil
}

func NewLatestStateReader(tx kv.Tx) state.StateReader {
	return state.NewReaderV4(tx.(kv.TemporalGetter))
}
func NewLatestStateWriter(txc wrap.TxContainer, blockNum uint64) state.StateWriter {
	domains := txc.Doms
	minTxNum, err := rawdbv3.TxNums.Min(domains.Tx(), blockNum)
	if err != nil {
		panic(err)
	}
	domains.SetTxNum(uint64(int(minTxNum) + /* 1 system txNum in beginning of block */ 1))
	return state.NewWriterV4(domains)
}

func CreateLatestCachedStateReader(cache kvcache.CacheView, tx kv.Tx) state.StateReader {
	return state.NewCachedReader3(cache, tx.(kv.TemporalTx))
}
