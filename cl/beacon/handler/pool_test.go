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

package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/log/v3"
	"github.com/ledgerwatch/erigon/cl/clparams"
	"github.com/ledgerwatch/erigon/cl/cltypes"
	"github.com/ledgerwatch/erigon/cl/cltypes/solid"
)

func TestPoolAttesterSlashings(t *testing.T) {
	attesterSlashing := &cltypes.AttesterSlashing{
		Attestation_1: &cltypes.IndexedAttestation{
			AttestingIndices: solid.NewRawUint64List(2048, []uint64{2, 3, 4, 5, 6}),
			Data:             solid.NewAttestationData(),
		},
		Attestation_2: &cltypes.IndexedAttestation{
			AttestingIndices: solid.NewRawUint64List(2048, []uint64{2, 3, 4, 1, 6}),
			Data:             solid.NewAttestationData(),
		},
	}
	// find server
	_, _, _, _, _, handler, _, _, _, _ := setupTestingHandler(t, clparams.Phase0Version, log.Root())

	server := httptest.NewServer(handler.mux)
	defer server.Close()
	// json
	req, err := json.Marshal(attesterSlashing)
	require.NoError(t, err)
	// post attester slashing
	resp, err := server.Client().Post(server.URL+"/eth/v1/beacon/pool/attester_slashings", "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	// get attester slashings
	resp, err = server.Client().Get(server.URL + "/eth/v1/beacon/pool/attester_slashings")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	out := struct {
		Data []*cltypes.AttesterSlashing `json:"data"`
	}{
		Data: []*cltypes.AttesterSlashing{
			cltypes.NewAttesterSlashing(),
		},
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)

	require.Equal(t, 1, len(out.Data))
	require.Equal(t, attesterSlashing, out.Data[0])
}

func TestPoolProposerSlashings(t *testing.T) {
	proposerSlashing := &cltypes.ProposerSlashing{
		Header1: &cltypes.SignedBeaconBlockHeader{
			Header: &cltypes.BeaconBlockHeader{
				Slot:          1,
				ProposerIndex: 3,
			},
		},
		Header2: &cltypes.SignedBeaconBlockHeader{
			Header: &cltypes.BeaconBlockHeader{
				Slot:          2,
				ProposerIndex: 4,
			},
		},
	}
	// find server
	_, _, _, _, _, handler, _, _, _, _ := setupTestingHandler(t, clparams.Phase0Version, log.Root())

	server := httptest.NewServer(handler.mux)
	defer server.Close()
	// json
	req, err := json.Marshal(proposerSlashing)
	require.NoError(t, err)

	// post attester slashing
	resp, err := server.Client().Post(server.URL+"/eth/v1/beacon/pool/proposer_slashings", "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	// get attester slashings
	resp, err = server.Client().Get(server.URL + "/eth/v1/beacon/pool/proposer_slashings")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	out := struct {
		Data []*cltypes.ProposerSlashing `json:"data"`
	}{
		Data: []*cltypes.ProposerSlashing{},
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)

	require.Equal(t, 1, len(out.Data))
	require.Equal(t, proposerSlashing, out.Data[0])
}

func TestPoolVoluntaryExits(t *testing.T) {
	voluntaryExit := &cltypes.SignedVoluntaryExit{
		VoluntaryExit: &cltypes.VoluntaryExit{
			Epoch:          1,
			ValidatorIndex: 3,
		},
	}
	// find server
	_, _, _, _, _, handler, _, _, _, _ := setupTestingHandler(t, clparams.Phase0Version, log.Root())

	server := httptest.NewServer(handler.mux)
	defer server.Close()
	// json
	req, err := json.Marshal(voluntaryExit)
	require.NoError(t, err)
	// post attester slashing
	resp, err := server.Client().Post(server.URL+"/eth/v1/beacon/pool/voluntary_exits", "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	// get attester slashings
	resp, err = server.Client().Get(server.URL + "/eth/v1/beacon/pool/voluntary_exits")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	out := struct {
		Data []*cltypes.SignedVoluntaryExit `json:"data"`
	}{
		Data: []*cltypes.SignedVoluntaryExit{},
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)

	require.Equal(t, 1, len(out.Data))
	require.Equal(t, voluntaryExit, out.Data[0])
}

func TestPoolBlsToExecutionChainges(t *testing.T) {
	msg := []*cltypes.SignedBLSToExecutionChange{
		{
			Message: &cltypes.BLSToExecutionChange{
				ValidatorIndex: 45,
			},
			Signature: libcommon.Bytes96{2},
		},
		{
			Message: &cltypes.BLSToExecutionChange{
				ValidatorIndex: 46,
			},
		},
	}
	// find server
	_, _, _, _, _, handler, _, _, _, _ := setupTestingHandler(t, clparams.Phase0Version, log.Root())

	server := httptest.NewServer(handler.mux)
	defer server.Close()
	// json
	req, err := json.Marshal(msg)
	require.NoError(t, err)
	// post attester slashing
	resp, err := server.Client().Post(server.URL+"/eth/v1/beacon/pool/bls_to_execution_changes", "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	// get attester slashings
	resp, err = server.Client().Get(server.URL + "/eth/v1/beacon/pool/bls_to_execution_changes")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	out := struct {
		Data []*cltypes.SignedBLSToExecutionChange `json:"data"`
	}{
		Data: []*cltypes.SignedBLSToExecutionChange{},
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)

	require.Equal(t, 2, len(out.Data))
	require.Equal(t, msg[0], out.Data[0])
	require.Equal(t, msg[1], out.Data[1])
}

func TestPoolAggregatesAndProofs(t *testing.T) {
	msg := []*cltypes.SignedAggregateAndProof{
		{
			Message: &cltypes.AggregateAndProof{
				Aggregate: solid.NewAttestionFromParameters([]byte{1, 2}, solid.NewAttestationData(), libcommon.Bytes96{3, 45, 6}),
			},
			Signature: libcommon.Bytes96{2},
		},
		{
			Message: &cltypes.AggregateAndProof{
				Aggregate: solid.NewAttestionFromParameters([]byte{1, 2, 5, 6}, solid.NewAttestationData(), libcommon.Bytes96{3, 0, 6}),
			},
			Signature: libcommon.Bytes96{2, 3, 5},
		},
	}
	// find server
	_, _, _, _, _, handler, _, _, _, _ := setupTestingHandler(t, clparams.Phase0Version, log.Root())

	server := httptest.NewServer(handler.mux)
	defer server.Close()
	// json
	req, err := json.Marshal(msg)
	require.NoError(t, err)
	// post attester slashing
	resp, err := server.Client().Post(server.URL+"/eth/v1/validator/aggregate_and_proofs", "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	// get attester slashings
	resp, err = server.Client().Get(server.URL + "/eth/v1/beacon/pool/attestations")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	out := struct {
		Data []*solid.Attestation `json:"data"`
	}{
		Data: []*solid.Attestation{},
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)

	require.Equal(t, 2, len(out.Data))
	require.Equal(t, msg[0].Message.Aggregate, out.Data[0])
	require.Equal(t, msg[1].Message.Aggregate, out.Data[1])
}

func TestPoolSyncCommittees(t *testing.T) {
	msgs := []*cltypes.SyncCommitteeMessage{
		{
			Slot:            1,
			BeaconBlockRoot: libcommon.Hash{1, 2, 3, 4, 5, 6, 7, 8},
			ValidatorIndex:  3,
		},
	}
	_, _, _, s, _, handler, _, sd, _, _ := setupTestingHandler(t, clparams.BellatrixVersion, log.Root())

	sd.OnHeadState(s)
	server := httptest.NewServer(handler.mux)
	defer server.Close()
	// json
	req, err := json.Marshal(msgs)
	require.NoError(t, err)
	// post attester slashing
	resp, err := server.Client().Post(server.URL+"/eth/v1/beacon/pool/sync_committees", "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	// get attester slashings
	resp, err = server.Client().Get(server.URL + "/eth/v1/validator/sync_committee_contribution?slot=1&subcommittee_index=0&beacon_block_root=0x0102030405060708000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	out := struct {
		Data *cltypes.Contribution `json:"data"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)

	require.Equal(t, out.Data, &cltypes.Contribution{
		Slot:              1,
		BeaconBlockRoot:   libcommon.Hash{1, 2, 3, 4, 5, 6, 7, 8},
		SubcommitteeIndex: 0,
		AggregationBits:   make([]byte, cltypes.SyncCommitteeAggregationBitsSize),
	})
}

func TestPoolSyncContributionAndProofs(t *testing.T) {
	aggrBits := make([]byte, cltypes.SyncCommitteeAggregationBitsSize)
	aggrBits[0] = 1
	msgs := []*cltypes.SignedContributionAndProof{
		{
			Message: &cltypes.ContributionAndProof{
				Contribution: &cltypes.Contribution{
					Slot:            1,
					BeaconBlockRoot: libcommon.Hash{1, 2, 3, 4, 5, 6, 7, 8},
					AggregationBits: aggrBits,
				},
			},
		},
	}
	_, _, _, s, _, handler, _, sd, _, _ := setupTestingHandler(t, clparams.BellatrixVersion, log.Root())

	sd.OnHeadState(s)
	server := httptest.NewServer(handler.mux)
	defer server.Close()
	// json
	req, err := json.Marshal(msgs)
	require.NoError(t, err)
	// post attester slashing
	resp, err := server.Client().Post(server.URL+"/eth/v1/validator/contribution_and_proofs", "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	// get attester slashings
	resp, err = server.Client().Get(server.URL + "/eth/v1/validator/sync_committee_contribution?slot=1&subcommittee_index=0&beacon_block_root=0x0102030405060708000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	out := struct {
		Data *cltypes.Contribution `json:"data"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)

	require.Equal(t, out.Data, &cltypes.Contribution{
		Slot:              1,
		BeaconBlockRoot:   libcommon.Hash{1, 2, 3, 4, 5, 6, 7, 8},
		SubcommitteeIndex: 0,
		AggregationBits:   aggrBits,
	})
}
