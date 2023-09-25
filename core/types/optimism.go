package types

import (
	"errors"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// Optimism tx type
//
// Ref: https://github.com/ethereum-optimism/op-geth
const (
	OPDepositTxType = 0x7e
)

type DepositTx struct {
	// SourceHash uniquely identifies the source of the deposit
	SourceHash common.Hash
	// From is exposed through the types.Signer, not through TxData
	From common.Address
	// nil means contract creation
	To *common.Address `rlp:"nil"`
	// Mint is minted on L2, locked on L1, nil if no minting.
	Mint *big.Int `rlp:"nil"`
	// Value is transferred from L2 balance, executed after Mint (if any)
	Value *big.Int
	// gas limit
	Gas uint64
	// Field indicating if this transaction is exempt from the L2 gas limit.
	IsSystemTransaction bool
	// Normal Tx data
	Data []byte
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *DepositTx) copy() TxData {
	cpy := &DepositTx{
		SourceHash:          tx.SourceHash,
		From:                tx.From,
		To:                  copyAddressPtr(tx.To),
		Mint:                nil,
		Value:               new(big.Int),
		Gas:                 tx.Gas,
		IsSystemTransaction: tx.IsSystemTransaction,
		Data:                common.CopyBytes(tx.Data),
	}
	if tx.Mint != nil {
		cpy.Mint = new(big.Int).Set(tx.Mint)
	}
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	return cpy
}

// accessors for innerTx.
func (tx *DepositTx) txType() byte              { return OPDepositTxType }
func (tx *DepositTx) chainID() *big.Int         { return common.Big0 }
func (tx *DepositTx) accessList() AccessList    { return nil }
func (tx *DepositTx) data() []byte              { return tx.Data }
func (tx *DepositTx) gas() uint64               { return tx.Gas }
func (tx *DepositTx) gasFeeCap() *big.Int       { return new(big.Int) }
func (tx *DepositTx) gasTipCap() *big.Int       { return new(big.Int) }
func (tx *DepositTx) gasPrice() *big.Int        { return new(big.Int) }
func (tx *DepositTx) value() *big.Int           { return tx.Value }
func (tx *DepositTx) nonce() uint64             { return 0 }
func (tx *DepositTx) to() *common.Address       { return tx.To }
func (tx *DepositTx) blobGas() uint64           { return 0 }
func (tx *DepositTx) blobGasFeeCap() *big.Int   { return nil }
func (tx *DepositTx) blobHashes() []common.Hash { return nil }

// func (tx *DepositTx) isSystemTx() bool          { return tx.IsSystemTransaction }

func (tx *DepositTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	return dst.Set(new(big.Int))
}

// func (tx *DepositTx) effectiveNonce() *uint64 { return nil }

func (tx *DepositTx) rawSignatureValues() (v, r, s *big.Int) {
	return common.Big0, common.Big0, common.Big0
}

func (tx *DepositTx) setSignatureValues(chainID, v, r, s *big.Int) {
	// this is a noop for deposit transactions
}

type depositTxWithNonce struct {
	DepositTx
	EffectiveNonce uint64
}

// EncodeRLP ensures that RLP encoding this transaction excludes the nonce. Otherwise, the tx Hash would change
func (tx *depositTxWithNonce) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, tx.DepositTx)
}

func (tx *Transaction) unmarshalOptimismJSON(dec txJSON) error {
	var inner TxData
	switch dec.Type {
	case OPDepositTxType:
		if dec.AccessList != nil || dec.MaxFeePerGas != nil ||
			dec.MaxPriorityFeePerGas != nil {
			return errors.New("unexpected field(s) in deposit transaction")
		}
		if dec.GasPrice != nil && dec.GasPrice.ToInt().Cmp(common.Big0) != 0 {
			return errors.New("deposit transaction GasPrice must be 0")
		}
		if (dec.V != nil && dec.V.ToInt().Cmp(common.Big0) != 0) ||
			(dec.R != nil && dec.R.ToInt().Cmp(common.Big0) != 0) ||
			(dec.S != nil && dec.S.ToInt().Cmp(common.Big0) != 0) {
			return errors.New("deposit transaction signature must be 0 or unset")
		}
		var itx DepositTx
		inner = &itx
		if dec.To != nil {
			itx.To = dec.To
		}
		if dec.Gas == nil {
			return errors.New("missing required field 'gas' for txdata")
		}
		itx.Gas = uint64(*dec.Gas)
		if dec.Value == nil {
			return errors.New("missing required field 'value' in transaction")
		}
		itx.Value = (*big.Int)(dec.Value)
		// mint may be omitted or nil if there is nothing to mint.
		itx.Mint = (*big.Int)(dec.Mint)
		if dec.Input == nil {
			return errors.New("missing required field 'input' in transaction")
		}
		itx.Data = *dec.Input
		if dec.From == nil {
			return errors.New("missing required field 'from' in transaction")
		}
		itx.From = *dec.From
		if dec.SourceHash == nil {
			return errors.New("missing required field 'sourceHash' in transaction")
		}
		itx.SourceHash = *dec.SourceHash
		// IsSystemTx may be omitted. Defaults to false.
		if dec.IsSystemTx != nil {
			itx.IsSystemTransaction = *dec.IsSystemTx
		}

		if dec.Nonce != nil {
			inner = &depositTxWithNonce{DepositTx: itx, EffectiveNonce: uint64(*dec.Nonce)}
		}
	default:
		return ErrTxTypeNotSupported
	}

	tx.setDecoded(inner, 0)

	return nil
}
