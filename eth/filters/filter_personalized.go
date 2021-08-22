package filters

import (
	"context"
	"fmt"
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
)

// --------------------------
// extending the filter_system.go
// --------------------------
// SubscribePendingTxs creates a subscription that writes transaction hashes for
// transactions that enter the transaction pool.
func (es *EventSystem) SubscribePendingTxsData(txs chan []*types.Transaction) *Subscription {
	sub := &subscription{
		id:        rpc.NewID(),
		typ:       PendingTransactionsDataSubscription,
		created:   time.Now(),
		logs:      make(chan []*types.Log),
		txs:       txs,
		headers:   make(chan *types.Header),
		installed: make(chan struct{}),
		err:       make(chan error),
	}
	return es.subscribe(sub)
}

// ----------------------
// new filter
// ----------------------
// newFilteredTransactions changes to filter
func (api *PublicFilterAPI) NewFilteredTransactions(ctx context.Context, txCriteria TransactionCriteria) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		transactions := make(chan []*types.Transaction, 128)
		pendingTxSub := api.events.SubscribePendingTxsData(transactions)

		for {
			select {
			case transactions := <-transactions:
				// a single tx hash in one notification.
				for _, t := range transactions {
					if txCriteria.check(t) {
						notifier.Notify(rpcSub.ID, t.Hash())
					}
				}
			case <-rpcSub.Err():
				pendingTxSub.Unsubscribe()
				return
			case <-notifier.Closed():
				pendingTxSub.Unsubscribe()
				return
			}
		}
	}()

	return rpcSub, nil
}

// ----------------------
// FunctionSelector type
// ----------------------
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

func CallDataToFunctionSelector(b []byte) (fs FunctionSelector) {
	if b == nil {
		return fs // with empty byte array
	}
	if len(b) > len(fs) {
		b = b[len(b)-FunctionSelectorLength:]
	}
	copy(fs[FunctionSelectorLength-len(b):], b)
	return
}

type TransactionCriteria struct {
	// calldata function selector
	Method FunctionSelector

	// TODO: not working, consider to implement
	WithEOA bool // when false filters out transactions targeting the external owned authorities
}

func (filter *TransactionCriteria) check(tx *types.Transaction) bool {
	// filtering by function selector
	if len(tx.Data()) >= 4 { // having calldata, first 4 bytes as a function selector
		var dataSlice []byte = tx.Data()[0:FunctionSelectorLength]
		dataSliceFunctionSelector := CallDataToFunctionSelector(dataSlice)
		fmt.Printf("4byte: %v -- bytes to hash: %v, function selctor hex: %v, data slice seelector: %v\n",
			len(dataSlice), common.BytesToHash(dataSlice), filter.Method.GetHex(), dataSliceFunctionSelector.GetHex())
		return filter.Method == dataSliceFunctionSelector
	} else {
		return len(filter.Method) == 0 // no filter method defined, then no criteria defined and it's ok to pass through
	}
}
