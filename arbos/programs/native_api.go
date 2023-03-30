// Copyright 2023, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE

//go:build !js
// +build !js

package programs

/*
#cgo CFLAGS: -g -Wall -I../../target/include/
#cgo LDFLAGS: ${SRCDIR}/../../target/lib/libstylus.a -ldl -lm
#include "arbitrator.h"

typedef uint32_t u32;
typedef uint64_t u64;
typedef size_t usize;

Bytes32 getBytes32Impl(usize api, Bytes32 key, u64 * cost);
Bytes32 getBytes32Wrap(usize api, Bytes32 key, u64 * cost) {
    return getBytes32Impl(api, key, cost);
}

GoApiStatus setBytes32Impl(usize api, Bytes32 key, Bytes32 value, u64 * cost, RustVec * error);
GoApiStatus setBytes32Wrap(usize api, Bytes32 key, Bytes32 value, u64 * cost, RustVec * error) {
    return setBytes32Impl(api, key, value, cost, error);
}

GoApiStatus contractCallImpl(usize api, Bytes20 contract, RustVec * calldata, u64 * gas, Bytes32 value, u32 * len);
GoApiStatus contractCallWrap(usize api, Bytes20 contract, RustVec * calldata, u64 * gas, Bytes32 value, u32 * len) {
    return contractCallImpl(api, contract, calldata, gas, value, len);
}

GoApiStatus delegateCallImpl(usize api, Bytes20 contract, RustVec * calldata, u64 * gas, u32 * len);
GoApiStatus delegateCallWrap(usize api, Bytes20 contract, RustVec * calldata, u64 * gas, u32 * len) {
    return delegateCallImpl(api, contract, calldata, gas, len);
}

GoApiStatus staticCallImpl(usize api, Bytes20 contract, RustVec * calldata, u64 * gas, u32 * len);
GoApiStatus staticCallWrap(usize api, Bytes20 contract, RustVec * calldata, u64 * gas, u32 * len) {
    return staticCallImpl(api, contract, calldata, gas, len);
}

void getReturnDataImpl(usize api, RustVec * data);
void getReturnDataWrap(usize api, RustVec * data) {
    return getReturnDataImpl(api, data);
}
*/
import "C"
import (
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

var apiClosures sync.Map
var apiIds int64 // atomic

type getBytes32Type func(key common.Hash) (value common.Hash, cost uint64)
type setBytes32Type func(key, value common.Hash) (cost uint64, err error)
type contractCallType func(
	contract common.Address, calldata []byte, gas uint64, value *big.Int) (
	retdata_len uint32, gas_left uint64, err error,
)
type delegateCallType func(
	contract common.Address, calldata []byte, gas uint64) (
	retdata_len uint32, gas_left uint64, err error,
)
type staticCallType func(
	contract common.Address, calldata []byte, gas uint64) (
	retdata_len uint32, gas_left uint64, err error,
)
type getReturnDataType func() []byte

type apiClosure struct {
	getBytes32    getBytes32Type
	setBytes32    setBytes32Type
	contractCall  contractCallType
	delegateCall  delegateCallType
	staticCall    staticCallType
	getReturnData getReturnDataType
}

func newAPI(
	getBytes32 getBytes32Type,
	setBytes32 setBytes32Type,
	contractCall contractCallType,
	delegateCall delegateCallType,
	staticCall staticCallType,
	getReturnData getReturnDataType,
) C.GoApi {
	id := atomic.AddInt64(&apiIds, 1)
	apiClosures.Store(id, apiClosure{
		getBytes32:    getBytes32,
		setBytes32:    setBytes32,
		contractCall:  contractCall,
		delegateCall:  delegateCall,
		staticCall:    staticCall,
		getReturnData: getReturnData,
	})
	return C.GoApi{
		get_bytes32:     (*[0]byte)(C.getBytes32Wrap),
		set_bytes32:     (*[0]byte)(C.setBytes32Wrap),
		contract_call:   (*[0]byte)(C.contractCallWrap),
		delegate_call:   (*[0]byte)(C.delegateCallWrap),
		static_call:     (*[0]byte)(C.staticCallWrap),
		get_return_data: (*[0]byte)(C.getReturnDataWrap),
		id:              u64(id),
	}
}

func getAPI(api usize) *apiClosure {
	any, ok := apiClosures.Load(int64(api))
	if !ok {
		log.Crit("failed to load stylus Go API", "id", api)
	}
	closures, ok := any.(apiClosure)
	if !ok {
		log.Crit("wrong type for stylus Go API", "id", api)
	}
	return &closures
}
