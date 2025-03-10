// Copyright 2019 The go-ethereum Authors
// (original work)
// Copyright 2024 The Erigon Authors
// (modifications)
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

//nolint:scopelint
package abi

import (
	"math/big"
	"reflect"
	"testing"

	libcommon "github.com/ledgerwatch/erigon-lib/common"

	"github.com/ledgerwatch/erigon/crypto"
)

func TestMakeTopics(t *testing.T) {
	type args struct {
		query [][]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    [][]libcommon.Hash
		wantErr bool
	}{
		{
			"support fixed byte types, right padded to 32 bytes",
			args{[][]interface{}{{[5]byte{1, 2, 3, 4, 5}}}},
			[][]libcommon.Hash{{libcommon.Hash{1, 2, 3, 4, 5}}},
			false,
		},
		{
			"support libcommon.Hash types in topics",
			args{[][]interface{}{{libcommon.Hash{1, 2, 3, 4, 5}}}},
			[][]libcommon.Hash{{libcommon.Hash{1, 2, 3, 4, 5}}},
			false,
		},
		{
			"support address types in topics",
			args{[][]interface{}{{libcommon.Address{1, 2, 3, 4, 5}}}},
			[][]libcommon.Hash{{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5}}},
			false,
		},
		{
			"support *big.Int types in topics",
			args{[][]interface{}{{big.NewInt(1).Lsh(big.NewInt(2), 254)}}},
			[][]libcommon.Hash{{libcommon.Hash{128}}},
			false,
		},
		{
			"support boolean types in topics",
			args{[][]interface{}{
				{true},
				{false},
			}},
			[][]libcommon.Hash{
				{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}},
				{libcommon.Hash{0}},
			},
			false,
		},
		{
			"support int/uint(8/16/32/64) types in topics",
			args{[][]interface{}{
				{int8(-2)},
				{int16(-3)},
				{int32(-4)},
				{int64(-5)},
				{int8(1)},
				{int16(256)},
				{int32(65536)},
				{int64(4294967296)},
				{uint8(1)},
				{uint16(256)},
				{uint32(65536)},
				{uint64(4294967296)},
			}},
			[][]libcommon.Hash{
				{libcommon.Hash{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 254}},
				{libcommon.Hash{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 253}},
				{libcommon.Hash{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 252}},
				{libcommon.Hash{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 251}},
				{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}},
				{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}},
				{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0}},
				{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0}},
				{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}},
				{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}},
				{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0}},
				{libcommon.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0}},
			},
			false,
		},
		{
			"support string types in topics",
			args{[][]interface{}{{"hello world"}}},
			[][]libcommon.Hash{{crypto.Keccak256Hash([]byte("hello world"))}},
			false,
		},
		{
			"support byte slice types in topics",
			args{[][]interface{}{{[]byte{1, 2, 3}}}},
			[][]libcommon.Hash{{crypto.Keccak256Hash([]byte{1, 2, 3})}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeTopics(tt.args.query...)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeTopics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeTopics() = %v, want %v", got, tt.want)
			}
		})
	}
}

type args struct {
	createObj func() interface{}
	resultObj func() interface{}
	resultMap func() map[string]interface{}
	fields    Arguments
	topics    []libcommon.Hash
}

type bytesStruct struct {
	StaticBytes [5]byte
}
type int8Struct struct {
	Int8Value int8
}
type int256Struct struct {
	Int256Value *big.Int
}

type hashStruct struct {
	HashValue libcommon.Hash
}

type funcStruct struct {
	FuncValue [24]byte
}

type topicTest struct {
	name    string
	args    args
	wantErr bool
}

func setupTopicsTests() []topicTest {
	bytesType, _ := NewType("bytes5", "", nil)
	int8Type, _ := NewType("int8", "", nil)
	int256Type, _ := NewType("int256", "", nil)
	tupleType, _ := NewType("tuple(int256,int8)", "", nil)
	stringType, _ := NewType("string", "", nil)
	funcType, _ := NewType("function", "", nil)

	tests := []topicTest{
		{
			name: "support fixed byte types, right padded to 32 bytes",
			args: args{
				createObj: func() interface{} { return &bytesStruct{} },
				resultObj: func() interface{} { return &bytesStruct{StaticBytes: [5]byte{1, 2, 3, 4, 5}} },
				resultMap: func() map[string]interface{} {
					return map[string]interface{}{"staticBytes": [5]byte{1, 2, 3, 4, 5}}
				},
				fields: Arguments{Argument{
					Name:    "staticBytes",
					Type:    bytesType,
					Indexed: true,
				}},
				topics: []libcommon.Hash{
					{1, 2, 3, 4, 5},
				},
			},
			wantErr: false,
		},
		{
			name: "int8 with negative value",
			args: args{
				createObj: func() interface{} { return &int8Struct{} },
				resultObj: func() interface{} { return &int8Struct{Int8Value: -1} },
				resultMap: func() map[string]interface{} {
					return map[string]interface{}{"int8Value": int8(-1)}
				},
				fields: Arguments{Argument{
					Name:    "int8Value",
					Type:    int8Type,
					Indexed: true,
				}},
				topics: []libcommon.Hash{
					{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
				},
			},
			wantErr: false,
		},
		{
			name: "int256 with negative value",
			args: args{
				createObj: func() interface{} { return &int256Struct{} },
				resultObj: func() interface{} { return &int256Struct{Int256Value: big.NewInt(-1)} },
				resultMap: func() map[string]interface{} {
					return map[string]interface{}{"int256Value": big.NewInt(-1)}
				},
				fields: Arguments{Argument{
					Name:    "int256Value",
					Type:    int256Type,
					Indexed: true,
				}},
				topics: []libcommon.Hash{
					{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
				},
			},
			wantErr: false,
		},
		{
			name: "hash type",
			args: args{
				createObj: func() interface{} { return &hashStruct{} },
				resultObj: func() interface{} { return &hashStruct{crypto.Keccak256Hash([]byte("stringtopic"))} },
				resultMap: func() map[string]interface{} {
					return map[string]interface{}{"hashValue": crypto.Keccak256Hash([]byte("stringtopic"))}
				},
				fields: Arguments{Argument{
					Name:    "hashValue",
					Type:    stringType,
					Indexed: true,
				}},
				topics: []libcommon.Hash{
					crypto.Keccak256Hash([]byte("stringtopic")),
				},
			},
			wantErr: false,
		},
		{
			name: "function type",
			args: args{
				createObj: func() interface{} { return &funcStruct{} },
				resultObj: func() interface{} {
					return &funcStruct{[24]byte{255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}}
				},
				resultMap: func() map[string]interface{} {
					return map[string]interface{}{"funcValue": [24]byte{255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}}
				},
				fields: Arguments{Argument{
					Name:    "funcValue",
					Type:    funcType,
					Indexed: true,
				}},
				topics: []libcommon.Hash{
					{0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
				},
			},
			wantErr: false,
		},
		{
			name: "error on topic/field count mismatch",
			args: args{
				createObj: func() interface{} { return nil },
				resultObj: func() interface{} { return nil },
				resultMap: func() map[string]interface{} { return make(map[string]interface{}) },
				fields: Arguments{Argument{
					Name:    "tupletype",
					Type:    tupleType,
					Indexed: true,
				}},
				topics: []libcommon.Hash{},
			},
			wantErr: true,
		},
		{
			name: "error on unindexed arguments",
			args: args{
				createObj: func() interface{} { return &int256Struct{} },
				resultObj: func() interface{} { return &int256Struct{} },
				resultMap: func() map[string]interface{} { return make(map[string]interface{}) },
				fields: Arguments{Argument{
					Name:    "int256Value",
					Type:    int256Type,
					Indexed: false,
				}},
				topics: []libcommon.Hash{
					{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
				},
			},
			wantErr: true,
		},
		{
			name: "error on tuple in topic reconstruction",
			args: args{
				createObj: func() interface{} { return &tupleType },
				resultObj: func() interface{} { return &tupleType },
				resultMap: func() map[string]interface{} { return make(map[string]interface{}) },
				fields: Arguments{Argument{
					Name:    "tupletype",
					Type:    tupleType,
					Indexed: true,
				}},
				topics: []libcommon.Hash{{0}},
			},
			wantErr: true,
		},
		{
			name: "error on improper encoded function",
			args: args{
				createObj: func() interface{} { return &funcStruct{} },
				resultObj: func() interface{} { return &funcStruct{} },
				resultMap: func() map[string]interface{} {
					return make(map[string]interface{})
				},
				fields: Arguments{Argument{
					Name:    "funcValue",
					Type:    funcType,
					Indexed: true,
				}},
				topics: []libcommon.Hash{
					{0, 0, 0, 0, 0, 0, 0, 128, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
				},
			},
			wantErr: true,
		},
	}

	return tests
}

func TestParseTopics(t *testing.T) {
	tests := setupTopicsTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createObj := tt.args.createObj()
			if err := ParseTopics(createObj, tt.args.fields, tt.args.topics); (err != nil) != tt.wantErr {
				t.Errorf("parseTopics() error = %v, wantErr %v", err, tt.wantErr)
			}
			resultObj := tt.args.resultObj()
			if !reflect.DeepEqual(createObj, resultObj) {
				t.Errorf("parseTopics() = %v, want %v", createObj, resultObj)
			}
		})
	}
}

func TestParseTopicsIntoMap(t *testing.T) {
	tests := setupTopicsTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outMap := make(map[string]interface{})
			if err := ParseTopicsIntoMap(outMap, tt.args.fields, tt.args.topics); (err != nil) != tt.wantErr {
				t.Errorf("parseTopicsIntoMap() error = %v, wantErr %v", err, tt.wantErr)
			}
			resultMap := tt.args.resultMap()
			if !reflect.DeepEqual(outMap, resultMap) {
				t.Errorf("parseTopicsIntoMap() = %v, want %v", outMap, resultMap)
			}
		})
	}
}
