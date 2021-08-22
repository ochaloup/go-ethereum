package filters

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	// the function selector of 4 bytes long (keccak256 hash)
	// see https://docs.soliditylang.org/en/v0.5.3/abi-spec.html#function-selector
	FunctionSelectorLength int = 4
	// filter subscription id, it can be used to know if it's the particular filter ID we want to work with
	FilteredTransactionsSubscription int = 2000
)

type FunctionSelector [FunctionSelectorLength]byte

// MarshalText returns the hex representation of a.
func (fs FunctionSelector) MarshalText() ([]byte, error) {
	return hexutil.Bytes(fs[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (fs *FunctionSelector) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("FunctionSelector", input, fs[:])
}

func (fs *FunctionSelector) GetHex() string {
	return common.Bytes2Hex(fs[:])
}

func BytesToFunctionSelector(b []byte) (fs FunctionSelector) {
	if b == nil {
		copy(fs[:], [4]byte)
	}
	if len(b) > len(fs) {
		b = b[len(b)-FunctionSelectorLength:]
	}
	copy(fs[FunctionSelectorLength-len(b):], b)
	return
}

type FilterMethod struct {
	Method FunctionSelector // filters by calldata function name
	// TODO: no functionality, consider to implement
	WithEOA bool // when false filters out transactions targeting the external owned authorities
}

// SubscribeFilteredTxs test
func (es *EventSystem) SubscribeFilteredTxs(hashes chan []common.Hash, methodFilter FilterMethod) *Subscription {
	sub := &subscription{
		id:        rpc.NewID(),
		typ:       FilteredTransactionsSubscription,
		txFilter:  methodFilter,
		created:   time.Now(),
		logs:      make(chan []*types.Log),
		hashes:    hashes,
		headers:   make(chan *types.Header),
		installed: make(chan struct{}),
		err:       make(chan error),
	}
	return es.subscribe(sub)
}
