package filters

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// the function selector of 4 bytes long (keccak256 hash)
// see https://docs.soliditylang.org/en/v0.5.3/abi-spec.html#function-selector
type FunctionSelector [4]byte

// MarshalText returns the hex representation of a.
func (fs FunctionSelector) MarshalText() ([]byte, error) {
	return hexutil.Bytes(fs[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (fs *FunctionSelector) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("FunctionSelector", input, fs[:])
}

type FilterMethod struct {
	Method  FunctionSelector // filters by calldata function name
	WithEOA bool             // when false filters out transactions targeting the external owned authorities
}
