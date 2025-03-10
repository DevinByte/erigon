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

package jsonrpc

import (
	"github.com/ledgerwatch/erigon-lib/log/v3"

	txpool "github.com/ledgerwatch/erigon-lib/gointerfaces/txpoolproto"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/kvcache"
	"github.com/ledgerwatch/erigon/cmd/rpcdaemon/cli/httpcfg"
	"github.com/ledgerwatch/erigon/consensus"
	"github.com/ledgerwatch/erigon/consensus/clique"
	"github.com/ledgerwatch/erigon/polygon/bor"
	"github.com/ledgerwatch/erigon/rpc"
	"github.com/ledgerwatch/erigon/turbo/rpchelper"
	"github.com/ledgerwatch/erigon/turbo/services"
)

// APIList describes the list of available RPC apis
func APIList(db kv.RoDB, eth rpchelper.ApiBackend, txPool txpool.TxpoolClient, mining txpool.MiningClient,
	filters *rpchelper.Filters, stateCache kvcache.Cache,
	blockReader services.FullBlockReader, cfg *httpcfg.HttpCfg, engine consensus.EngineReader,
	logger log.Logger,
) (list []rpc.API) {
	base := NewBaseApi(filters, stateCache, blockReader, cfg.WithDatadir, cfg.EvmCallTimeout, engine)
	ethImpl := NewEthAPI(base, db, eth, txPool, mining, cfg.Gascap, cfg.Feecap, cfg.ReturnDataLimit, cfg.AllowUnprotectedTxs, cfg.MaxGetProofRewindBlockCount, cfg.WebsocketSubscribeLogsChannelSize, logger)
	erigonImpl := NewErigonAPI(base, db, eth)
	txpoolImpl := NewTxPoolAPI(base, db, txPool)
	netImpl := NewNetAPIImpl(eth)
	debugImpl := NewPrivateDebugAPI(base, db, cfg.Gascap)
	traceImpl := NewTraceAPI(base, db, cfg)
	web3Impl := NewWeb3APIImpl(eth)
	dbImpl := NewDBAPIImpl() /* deprecated */
	adminImpl := NewAdminAPI(eth)
	parityImpl := NewParityAPIImpl(base, db)

	var borImpl *BorImpl

	type lazy interface {
		HasEngine() bool
		Engine() consensus.EngineReader
	}

	switch engine := engine.(type) {
	case *bor.Bor:
		borImpl = NewBorAPI(base, db)
	case lazy:
		if _, ok := engine.Engine().(*bor.Bor); !engine.HasEngine() || ok {
			borImpl = NewBorAPI(base, db)
		}
	}

	otsImpl := NewOtterscanAPI(base, db, cfg.OtsMaxPageSize)
	gqlImpl := NewGraphQLAPI(base, db)
	overlayImpl := NewOverlayAPI(base, db, cfg.Gascap, cfg.OverlayGetLogsTimeout, cfg.OverlayReplayBlockTimeout, otsImpl)

	if cfg.GraphQLEnabled {
		list = append(list, rpc.API{
			Namespace: "graphql",
			Public:    true,
			Service:   GraphQLAPI(gqlImpl),
			Version:   "1.0",
		})
	}

	for _, enabledAPI := range cfg.API {
		switch enabledAPI {
		case "eth":
			list = append(list, rpc.API{
				Namespace: "eth",
				Public:    true,
				Service:   EthAPI(ethImpl),
				Version:   "1.0",
			})
		case "debug":
			list = append(list, rpc.API{
				Namespace: "debug",
				Public:    true,
				Service:   PrivateDebugAPI(debugImpl),
				Version:   "1.0",
			})
		case "net":
			list = append(list, rpc.API{
				Namespace: "net",
				Public:    true,
				Service:   NetAPI(netImpl),
				Version:   "1.0",
			})
		case "txpool":
			list = append(list, rpc.API{
				Namespace: "txpool",
				Public:    true,
				Service:   TxPoolAPI(txpoolImpl),
				Version:   "1.0",
			})
		case "web3":
			list = append(list, rpc.API{
				Namespace: "web3",
				Public:    true,
				Service:   Web3API(web3Impl),
				Version:   "1.0",
			})
		case "trace":
			list = append(list, rpc.API{
				Namespace: "trace",
				Public:    true,
				Service:   TraceAPI(traceImpl),
				Version:   "1.0",
			})
		case "db": /* Deprecated */
			list = append(list, rpc.API{
				Namespace: "db",
				Public:    true,
				Service:   DBAPI(dbImpl),
				Version:   "1.0",
			})
		case "erigon":
			list = append(list, rpc.API{
				Namespace: "erigon",
				Public:    true,
				Service:   ErigonAPI(erigonImpl),
				Version:   "1.0",
			})
		case "bor":
			if borImpl != nil {
				list = append(list, rpc.API{
					Namespace: "bor",
					Public:    true,
					Service:   BorAPI(borImpl),
					Version:   "1.0",
				})
			}
		case "admin":
			list = append(list, rpc.API{
				Namespace: "admin",
				Public:    false,
				Service:   AdminAPI(adminImpl),
				Version:   "1.0",
			})
		case "parity":
			list = append(list, rpc.API{
				Namespace: "parity",
				Public:    false,
				Service:   ParityAPI(parityImpl),
				Version:   "1.0",
			})
		case "ots":
			list = append(list, rpc.API{
				Namespace: "ots",
				Public:    true,
				Service:   OtterscanAPI(otsImpl),
				Version:   "1.0",
			})
		case "clique":
			list = append(list, clique.NewCliqueAPI(db, engine, blockReader))
		case "overlay":
			list = append(list, rpc.API{
				Namespace: "overlay",
				Public:    true,
				Service:   OverlayAPI(overlayImpl),
				Version:   "1.0",
			})
		}
	}

	return list
}
