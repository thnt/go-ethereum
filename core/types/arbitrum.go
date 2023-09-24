package types

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/rlp"
)

// Arbitrum tx type
//
// Ref: https://github.com/OffchainLabs/go-ethereum/
const (
	ArbitrumDepositTxType         = 0x64
	ArbitrumUnsignedTxType        = 0x65
	ArbitrumContractTxType        = 0x66
	ArbitrumRetryTxType           = 0x68
	ArbitrumSubmitRetryableTxType = 0x69
	ArbitrumInternalTxType        = 0x6A
	ArbitrumLegacyTxType          = 0x78
)

var bigZero = big.NewInt(0)

var (
	arbosAddress = common.HexToAddress("0xa4b05")
	// arbSysAddress             = common.HexToAddress("0x64")
	// arbGasInfoAddress         = common.HexToAddress("0x6c")
	arbRetryableTxAddress = common.HexToAddress("0x6e")
	// nodeInterfaceAddress      = common.HexToAddress("0xc8")
	// nodeInterfaceDebugAddress = common.HexToAddress("0xc9")
)

type ArbitrumLegacyTxData struct {
	LegacyTx
	HashOverride      common.Hash // Hash cannot be locally computed from other fields
	EffectiveGasPrice uint64
	L1BlockNumber     uint64
	Sender            *common.Address `rlp:"optional,nil"` // only used in unsigned Txs
}

func (tx *ArbitrumLegacyTxData) copy() TxData {
	legacyCopy := tx.LegacyTx.copy().(*LegacyTx)
	var sender *common.Address
	if tx.Sender != nil {
		sender = new(common.Address)
		*sender = *tx.Sender
	}
	return &ArbitrumLegacyTxData{
		LegacyTx:          *legacyCopy,
		HashOverride:      tx.HashOverride,
		EffectiveGasPrice: tx.EffectiveGasPrice,
		L1BlockNumber:     tx.L1BlockNumber,
		Sender:            sender,
	}
}

// func NewArbitrumLegacyTx(origTx *Transaction, hashOverride common.Hash, effectiveGas uint64, l1Block uint64, senderOverride *common.Address) (*Transaction, error) {
// 	if origTx.Type() != LegacyTxType {
// 		return nil, errors.New("attempt to arbitrum-wrap non-legacy transaction")
// 	}
// 	legacyPtr := origTx.GetInner().(*LegacyTx)
// 	inner := ArbitrumLegacyTxData{
// 		LegacyTx:          *legacyPtr,
// 		HashOverride:      hashOverride,
// 		EffectiveGasPrice: effectiveGas,
// 		L1BlockNumber:     l1Block,
// 		Sender:            senderOverride,
// 	}
// 	return NewTx(&inner), nil
// }

func (tx *ArbitrumLegacyTxData) txType() byte { return ArbitrumLegacyTxType }

func (tx *ArbitrumLegacyTxData) EncodeOnlyLegacyInto(w *bytes.Buffer) {
	rlp.Encode(w, tx.LegacyTx)
}

type ArbitrumUnsignedTx struct {
	ChainId *big.Int
	From    common.Address

	Nonce     uint64          // nonce of sender account
	GasFeeCap *big.Int        // wei per gas
	Gas       uint64          // gas limit
	To        *common.Address `rlp:"nil"` // nil means contract creation
	Value     *big.Int        // wei amount
	Data      []byte          // contract invocation input data
}

func (tx *ArbitrumUnsignedTx) txType() byte { return ArbitrumUnsignedTxType }

func (tx *ArbitrumUnsignedTx) copy() TxData {
	cpy := &ArbitrumUnsignedTx{
		ChainId:   new(big.Int),
		Nonce:     tx.Nonce,
		GasFeeCap: new(big.Int),
		Gas:       tx.Gas,
		From:      tx.From,
		To:        nil,
		Value:     new(big.Int),
		Data:      common.CopyBytes(tx.Data),
	}
	if tx.ChainId != nil {
		cpy.ChainId.Set(tx.ChainId)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.To != nil {
		tmp := *tx.To
		cpy.To = &tmp
	}
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	return cpy
}

func (tx *ArbitrumUnsignedTx) chainID() *big.Int         { return tx.ChainId }
func (tx *ArbitrumUnsignedTx) accessList() AccessList    { return nil }
func (tx *ArbitrumUnsignedTx) data() []byte              { return tx.Data }
func (tx *ArbitrumUnsignedTx) gas() uint64               { return tx.Gas }
func (tx *ArbitrumUnsignedTx) gasPrice() *big.Int        { return tx.GasFeeCap }
func (tx *ArbitrumUnsignedTx) gasTipCap() *big.Int       { return bigZero }
func (tx *ArbitrumUnsignedTx) gasFeeCap() *big.Int       { return tx.GasFeeCap }
func (tx *ArbitrumUnsignedTx) value() *big.Int           { return tx.Value }
func (tx *ArbitrumUnsignedTx) nonce() uint64             { return tx.Nonce }
func (tx *ArbitrumUnsignedTx) to() *common.Address       { return tx.To }
func (tx *ArbitrumUnsignedTx) blobGas() uint64           { return 0 }
func (tx *ArbitrumUnsignedTx) blobGasFeeCap() *big.Int   { return nil }
func (tx *ArbitrumUnsignedTx) blobHashes() []common.Hash { return nil }

func (tx *ArbitrumUnsignedTx) rawSignatureValues() (v, r, s *big.Int) {
	return bigZero, bigZero, bigZero
}

func (tx *ArbitrumUnsignedTx) setSignatureValues(chainID, v, r, s *big.Int) {
}

func (tx *ArbitrumUnsignedTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	return dst.Set(baseFee)
}

// func (tx *ArbitrumUnsignedTx) encode(*bytes.Buffer) error {
// 	return errors.New("ArbitrumUnsignedTx not support encode method")
// }

// func (tx *ArbitrumUnsignedTx) decode([]byte) error {
// 	return errors.New("ArbitrumUnsignedTx not support decode method")
// }

type ArbitrumInternalTx struct {
	ChainId *big.Int
	Data    []byte
}

func (tx *ArbitrumInternalTx) txType() byte {
	return ArbitrumInternalTxType
}

func (tx *ArbitrumInternalTx) copy() TxData {
	return &ArbitrumInternalTx{
		new(big.Int).Set(tx.ChainId),
		common.CopyBytes(tx.Data),
	}
}

func (tx *ArbitrumInternalTx) chainID() *big.Int         { return tx.ChainId }
func (tx *ArbitrumInternalTx) accessList() AccessList    { return nil }
func (tx *ArbitrumInternalTx) data() []byte              { return tx.Data }
func (tx *ArbitrumInternalTx) gas() uint64               { return 0 }
func (tx *ArbitrumInternalTx) gasPrice() *big.Int        { return bigZero }
func (tx *ArbitrumInternalTx) gasTipCap() *big.Int       { return bigZero }
func (tx *ArbitrumInternalTx) gasFeeCap() *big.Int       { return bigZero }
func (tx *ArbitrumInternalTx) value() *big.Int           { return common.Big0 }
func (tx *ArbitrumInternalTx) nonce() uint64             { return 0 }
func (tx *ArbitrumInternalTx) to() *common.Address       { return &arbosAddress }
func (tx *ArbitrumInternalTx) blobGas() uint64           { return 0 }
func (tx *ArbitrumInternalTx) blobGasFeeCap() *big.Int   { return nil }
func (tx *ArbitrumInternalTx) blobHashes() []common.Hash { return nil }

func (tx *ArbitrumInternalTx) rawSignatureValues() (v, r, s *big.Int) {
	return bigZero, bigZero, bigZero
}

func (tx *ArbitrumInternalTx) setSignatureValues(chainID, v, r, s *big.Int) {
}

func (tx *ArbitrumInternalTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	return dst.Set(bigZero)
}

// func (tx *ArbitrumInternalTx) encode(*bytes.Buffer) error {
// 	return errors.New("ArbitrumInternalTx not support encode method")
// }

// func (tx *ArbitrumInternalTx) decode([]byte) error {
// 	return errors.New("ArbitrumInternalTx not support decode method")
// }

type ArbitrumDepositTx struct {
	ChainId     *big.Int
	L1RequestId common.Hash
	From        common.Address
	To          common.Address
	Value       *big.Int
}

func (tx *ArbitrumDepositTx) txType() byte {
	return ArbitrumDepositTxType
}

func (tx *ArbitrumDepositTx) copy() TxData {
	dtx := &ArbitrumDepositTx{
		ChainId:     new(big.Int),
		L1RequestId: tx.L1RequestId,
		From:        tx.From,
		To:          tx.To,
		Value:       new(big.Int),
	}
	if tx.ChainId != nil {
		dtx.ChainId.Set(tx.ChainId)
	}
	if tx.Value != nil {
		dtx.Value.Set(tx.Value)
	}
	return dtx
}

func (tx *ArbitrumDepositTx) chainID() *big.Int         { return tx.ChainId }
func (tx *ArbitrumDepositTx) accessList() AccessList    { return nil }
func (tx *ArbitrumDepositTx) data() []byte              { return nil }
func (tx *ArbitrumDepositTx) gas() uint64               { return 0 }
func (tx *ArbitrumDepositTx) gasPrice() *big.Int        { return bigZero }
func (tx *ArbitrumDepositTx) gasTipCap() *big.Int       { return bigZero }
func (tx *ArbitrumDepositTx) gasFeeCap() *big.Int       { return bigZero }
func (tx *ArbitrumDepositTx) value() *big.Int           { return tx.Value }
func (tx *ArbitrumDepositTx) nonce() uint64             { return 0 }
func (tx *ArbitrumDepositTx) to() *common.Address       { return &tx.To }
func (tx *ArbitrumDepositTx) blobGas() uint64           { return 0 }
func (tx *ArbitrumDepositTx) blobGasFeeCap() *big.Int   { return nil }
func (tx *ArbitrumDepositTx) blobHashes() []common.Hash { return nil }

func (tx *ArbitrumDepositTx) rawSignatureValues() (v, r, s *big.Int) {
	return bigZero, bigZero, bigZero
}

func (tx *ArbitrumDepositTx) setSignatureValues(chainID, v, r, s *big.Int) {
}

func (tx *ArbitrumDepositTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	return dst.Set(bigZero)
}

// func (tx *ArbitrumDepositTx) encode(*bytes.Buffer) error {
// 	return errors.New("ArbitrumDepositTx not support encode method")
// }

// func (tx *ArbitrumDepositTx) decode([]byte) error {
// 	return errors.New("ArbitrumDepositTx not support decode method")
// }

type ArbitrumContractTx struct {
	ChainId   *big.Int
	RequestId common.Hash
	From      common.Address

	GasFeeCap *big.Int        // wei per gas
	Gas       uint64          // gas limit
	To        *common.Address `rlp:"nil"` // nil means contract creation
	Value     *big.Int        // wei amount
	Data      []byte          // contract invocation input data
}

func (tx *ArbitrumContractTx) txType() byte { return ArbitrumContractTxType }

func (tx *ArbitrumContractTx) copy() TxData {
	cpy := &ArbitrumContractTx{
		ChainId:   new(big.Int),
		RequestId: tx.RequestId,
		GasFeeCap: new(big.Int),
		Gas:       tx.Gas,
		From:      tx.From,
		To:        nil,
		Value:     new(big.Int),
		Data:      common.CopyBytes(tx.Data),
	}
	if tx.ChainId != nil {
		cpy.ChainId.Set(tx.ChainId)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.To != nil {
		tmp := *tx.To
		cpy.To = &tmp
	}
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	return cpy
}

func (tx *ArbitrumContractTx) chainID() *big.Int         { return tx.ChainId }
func (tx *ArbitrumContractTx) accessList() AccessList    { return nil }
func (tx *ArbitrumContractTx) data() []byte              { return tx.Data }
func (tx *ArbitrumContractTx) gas() uint64               { return tx.Gas }
func (tx *ArbitrumContractTx) gasPrice() *big.Int        { return tx.GasFeeCap }
func (tx *ArbitrumContractTx) gasTipCap() *big.Int       { return bigZero }
func (tx *ArbitrumContractTx) gasFeeCap() *big.Int       { return tx.GasFeeCap }
func (tx *ArbitrumContractTx) value() *big.Int           { return tx.Value }
func (tx *ArbitrumContractTx) nonce() uint64             { return 0 }
func (tx *ArbitrumContractTx) to() *common.Address       { return tx.To }
func (tx *ArbitrumContractTx) blobGas() uint64           { return 0 }
func (tx *ArbitrumContractTx) blobGasFeeCap() *big.Int   { return nil }
func (tx *ArbitrumContractTx) blobHashes() []common.Hash { return nil }

func (tx *ArbitrumContractTx) rawSignatureValues() (v, r, s *big.Int) {
	return bigZero, bigZero, bigZero
}
func (tx *ArbitrumContractTx) setSignatureValues(chainID, v, r, s *big.Int) {}

func (tx *ArbitrumContractTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	return dst.Set(baseFee)
}

// func (tx *ArbitrumContractTx) encode(*bytes.Buffer) error {
// 	return errors.New("ArbitrumContractTx not support encode method")
// }

// func (tx *ArbitrumContractTx) decode([]byte) error {
// 	return errors.New("ArbitrumContractTx not support decode method")
// }

type ArbitrumRetryTx struct {
	ChainId *big.Int
	Nonce   uint64
	From    common.Address

	GasFeeCap           *big.Int        // wei per gas
	Gas                 uint64          // gas limit
	To                  *common.Address `rlp:"nil"` // nil means contract creation
	Value               *big.Int        // wei amount
	Data                []byte          // contract invocation input data
	TicketId            common.Hash
	RefundTo            common.Address
	MaxRefund           *big.Int // the maximum refund sent to RefundTo (the rest goes to From)
	SubmissionFeeRefund *big.Int // the submission fee to refund if successful (capped by MaxRefund)
}

func (tx *ArbitrumRetryTx) txType() byte { return ArbitrumRetryTxType }

func (tx *ArbitrumRetryTx) copy() TxData {
	cpy := &ArbitrumRetryTx{
		ChainId:             new(big.Int),
		Nonce:               tx.Nonce,
		GasFeeCap:           new(big.Int),
		Gas:                 tx.Gas,
		From:                tx.From,
		To:                  nil,
		Value:               new(big.Int),
		Data:                common.CopyBytes(tx.Data),
		TicketId:            tx.TicketId,
		RefundTo:            tx.RefundTo,
		MaxRefund:           new(big.Int),
		SubmissionFeeRefund: new(big.Int),
	}
	if tx.ChainId != nil {
		cpy.ChainId.Set(tx.ChainId)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.To != nil {
		tmp := *tx.To
		cpy.To = &tmp
	}
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.MaxRefund != nil {
		cpy.MaxRefund.Set(tx.MaxRefund)
	}
	if tx.SubmissionFeeRefund != nil {
		cpy.SubmissionFeeRefund.Set(tx.SubmissionFeeRefund)
	}
	return cpy
}

func (tx *ArbitrumRetryTx) chainID() *big.Int         { return tx.ChainId }
func (tx *ArbitrumRetryTx) accessList() AccessList    { return nil }
func (tx *ArbitrumRetryTx) data() []byte              { return tx.Data }
func (tx *ArbitrumRetryTx) gas() uint64               { return tx.Gas }
func (tx *ArbitrumRetryTx) gasPrice() *big.Int        { return tx.GasFeeCap }
func (tx *ArbitrumRetryTx) gasTipCap() *big.Int       { return bigZero }
func (tx *ArbitrumRetryTx) gasFeeCap() *big.Int       { return tx.GasFeeCap }
func (tx *ArbitrumRetryTx) value() *big.Int           { return tx.Value }
func (tx *ArbitrumRetryTx) nonce() uint64             { return tx.Nonce }
func (tx *ArbitrumRetryTx) to() *common.Address       { return tx.To }
func (tx *ArbitrumRetryTx) blobGas() uint64           { return 0 }
func (tx *ArbitrumRetryTx) blobGasFeeCap() *big.Int   { return nil }
func (tx *ArbitrumRetryTx) blobHashes() []common.Hash { return nil }

func (tx *ArbitrumRetryTx) rawSignatureValues() (v, r, s *big.Int) {
	return bigZero, bigZero, bigZero
}
func (tx *ArbitrumRetryTx) setSignatureValues(chainID, v, r, s *big.Int) {}

func (tx *ArbitrumRetryTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	return dst.Set(baseFee)
}

// func (tx *ArbitrumRetryTx) encode(*bytes.Buffer) error {
// 	return errors.New("ArbitrumRetryTx not support encode method")
// }

// func (tx *ArbitrumRetryTx) decode([]byte) error {
// 	return errors.New("ArbitrumRetryTx not support decode method")
// }

type ArbitrumSubmitRetryableTx struct {
	ChainId   *big.Int
	RequestId common.Hash
	From      common.Address
	L1BaseFee *big.Int

	DepositValue     *big.Int
	GasFeeCap        *big.Int        // wei per gas
	Gas              uint64          // gas limit
	RetryTo          *common.Address `rlp:"nil"` // nil means contract creation
	RetryValue       *big.Int        // wei amount
	Beneficiary      common.Address
	MaxSubmissionFee *big.Int
	FeeRefundAddr    common.Address
	RetryData        []byte // contract invocation input data
}

func (tx *ArbitrumSubmitRetryableTx) txType() byte { return ArbitrumSubmitRetryableTxType }

func (tx *ArbitrumSubmitRetryableTx) copy() TxData {
	cpy := &ArbitrumSubmitRetryableTx{
		ChainId:          new(big.Int),
		RequestId:        tx.RequestId,
		DepositValue:     new(big.Int),
		L1BaseFee:        new(big.Int),
		GasFeeCap:        new(big.Int),
		Gas:              tx.Gas,
		From:             tx.From,
		RetryTo:          tx.RetryTo,
		RetryValue:       new(big.Int),
		Beneficiary:      tx.Beneficiary,
		MaxSubmissionFee: new(big.Int),
		FeeRefundAddr:    tx.FeeRefundAddr,
		RetryData:        common.CopyBytes(tx.RetryData),
	}
	if tx.ChainId != nil {
		cpy.ChainId.Set(tx.ChainId)
	}
	if tx.DepositValue != nil {
		cpy.DepositValue.Set(tx.DepositValue)
	}
	if tx.L1BaseFee != nil {
		cpy.L1BaseFee.Set(tx.L1BaseFee)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.RetryTo != nil {
		tmp := *tx.RetryTo
		cpy.RetryTo = &tmp
	}
	if tx.RetryValue != nil {
		cpy.RetryValue.Set(tx.RetryValue)
	}
	if tx.MaxSubmissionFee != nil {
		cpy.MaxSubmissionFee.Set(tx.MaxSubmissionFee)
	}
	return cpy
}

func (tx *ArbitrumSubmitRetryableTx) chainID() *big.Int         { return tx.ChainId }
func (tx *ArbitrumSubmitRetryableTx) accessList() AccessList    { return nil }
func (tx *ArbitrumSubmitRetryableTx) gas() uint64               { return tx.Gas }
func (tx *ArbitrumSubmitRetryableTx) gasPrice() *big.Int        { return tx.GasFeeCap }
func (tx *ArbitrumSubmitRetryableTx) gasTipCap() *big.Int       { return big.NewInt(0) }
func (tx *ArbitrumSubmitRetryableTx) gasFeeCap() *big.Int       { return tx.GasFeeCap }
func (tx *ArbitrumSubmitRetryableTx) value() *big.Int           { return common.Big0 }
func (tx *ArbitrumSubmitRetryableTx) nonce() uint64             { return 0 }
func (tx *ArbitrumSubmitRetryableTx) to() *common.Address       { return &arbRetryableTxAddress }
func (tx *ArbitrumSubmitRetryableTx) blobGas() uint64           { return 0 }
func (tx *ArbitrumSubmitRetryableTx) blobGasFeeCap() *big.Int   { return nil }
func (tx *ArbitrumSubmitRetryableTx) blobHashes() []common.Hash { return nil }

func (tx *ArbitrumSubmitRetryableTx) rawSignatureValues() (v, r, s *big.Int) {
	return bigZero, bigZero, bigZero
}
func (tx *ArbitrumSubmitRetryableTx) setSignatureValues(chainID, v, r, s *big.Int) {}

func (tx *ArbitrumSubmitRetryableTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	return dst.Set(baseFee)
}

func (tx *ArbitrumSubmitRetryableTx) data() []byte {
	var retryTo common.Address
	if tx.RetryTo != nil {
		retryTo = *tx.RetryTo
	}
	data := make([]byte, 0)
	data = append(data, tx.RequestId.Bytes()...)
	data = append(data, math.U256Bytes(tx.L1BaseFee)...)
	data = append(data, math.U256Bytes(tx.DepositValue)...)
	data = append(data, math.U256Bytes(tx.RetryValue)...)
	data = append(data, math.U256Bytes(tx.GasFeeCap)...)
	data = append(data, math.U256Bytes(new(big.Int).SetUint64(tx.Gas))...)
	data = append(data, math.U256Bytes(tx.MaxSubmissionFee)...)
	data = append(data, make([]byte, 12)...)
	data = append(data, tx.FeeRefundAddr.Bytes()...)
	data = append(data, make([]byte, 12)...)
	data = append(data, tx.Beneficiary.Bytes()...)
	data = append(data, make([]byte, 12)...)
	data = append(data, retryTo.Bytes()...)
	offset := len(data) + 32
	data = append(data, math.U256Bytes(big.NewInt(int64(offset)))...)
	data = append(data, math.U256Bytes(big.NewInt(int64(len(tx.RetryData))))...)
	data = append(data, tx.RetryData...)
	extra := len(tx.RetryData) % 32
	if extra > 0 {
		data = append(data, make([]byte, 32-extra)...)
	}
	data = append(hexutil.MustDecode("0xc9f95d32"), data...)
	return data
}

// func (tx *ArbitrumSubmitRetryableTx) encode(*bytes.Buffer) error {
// 	return errors.New("ArbitrumSubmitRetryableTx not support encode method")
// }

// func (tx *ArbitrumSubmitRetryableTx) decode([]byte) error {
// 	return errors.New("ArbitrumSubmitRetryableTx not support decode method")
// }

// func (tx *Transaction) GetInner() TxData {
// 	return tx.inner.copy()
// }

func (tx *Transaction) unmarshalArbitrumJSON(dec txJSON) error {
	var inner TxData

	switch dec.Type {

	case ArbitrumLegacyTxType:
		var itx LegacyTx
		if dec.To != nil {
			itx.To = dec.To
		}
		if dec.Nonce == nil {
			return errors.New("missing required field 'nonce' in transaction")
		}
		itx.Nonce = uint64(*dec.Nonce)
		if dec.GasPrice == nil {
			return errors.New("missing required field 'gasPrice' in transaction")
		}
		itx.GasPrice = (*big.Int)(dec.GasPrice)
		if dec.Gas == nil {
			return errors.New("missing required field 'gas' in transaction")
		}
		itx.Gas = uint64(*dec.Gas)
		if dec.Value == nil {
			return errors.New("missing required field 'value' in transaction")
		}
		itx.Value = (*big.Int)(dec.Value)
		if dec.Input == nil {
			return errors.New("missing required field 'input' in transaction")
		}
		itx.Data = *dec.Input
		if dec.V == nil {
			return errors.New("missing required field 'v' in transaction")
		}
		itx.V = (*big.Int)(dec.V)
		if dec.R == nil {
			return errors.New("missing required field 'r' in transaction")
		}
		itx.R = (*big.Int)(dec.R)
		if dec.S == nil {
			return errors.New("missing required field 's' in transaction")
		}
		itx.S = (*big.Int)(dec.S)
		withSignature := itx.V.Sign() != 0 || itx.R.Sign() != 0 || itx.S.Sign() != 0
		if withSignature {
			if err := sanityCheckSignature(itx.V, itx.R, itx.S, true); err != nil {
				return err
			}
		}
		if dec.EffectiveGasPrice == nil {
			return errors.New("missing required field 'EffectiveGasPrice' in transaction")
		}
		if dec.L1BlockNumber == nil {
			return errors.New("missing required field 'L1BlockNumber' in transaction")
		}
		inner = &ArbitrumLegacyTxData{
			LegacyTx:          itx,
			HashOverride:      dec.Hash,
			EffectiveGasPrice: uint64(*dec.EffectiveGasPrice),
			L1BlockNumber:     uint64(*dec.L1BlockNumber),
			Sender:            dec.From,
		}

	case ArbitrumInternalTxType:
		if dec.ChainID == nil {
			return errors.New("missing required field 'chainId' in transaction")
		}
		if dec.Input == nil {
			return errors.New("missing required field 'input' in transaction")
		}
		inner = &ArbitrumInternalTx{
			ChainId: (*big.Int)(dec.ChainID),
			Data:    *dec.Input,
		}

	case ArbitrumDepositTxType:
		if dec.ChainID == nil {
			return errors.New("missing required field 'chainId' in transaction")
		}
		if dec.RequestId == nil {
			return errors.New("missing required field 'requestId' in transaction")
		}
		if dec.To == nil {
			return errors.New("missing required field 'to' in transaction")
		}
		if dec.From == nil {
			return errors.New("missing required field 'from' in transaction")
		}
		if dec.Value == nil {
			return errors.New("missing required field 'value' in transaction")
		}
		inner = &ArbitrumDepositTx{
			ChainId:     (*big.Int)(dec.ChainID),
			L1RequestId: *dec.RequestId,
			To:          *dec.To,
			From:        *dec.From,
			Value:       (*big.Int)(dec.Value),
		}

	case ArbitrumUnsignedTxType:
		if dec.ChainID == nil {
			return errors.New("missing required field 'chainId' in transaction")
		}
		if dec.From == nil {
			return errors.New("missing required field 'from' in transaction")
		}
		if dec.Nonce == nil {
			return errors.New("missing required field 'nonce' in transaction")
		}
		if dec.MaxFeePerGas == nil {
			return errors.New("missing required field 'maxFeePerGas' for txdata")
		}
		if dec.Gas == nil {
			return errors.New("missing required field 'gas' in txdata")
		}
		if dec.Value == nil {
			return errors.New("missing required field 'value' in transaction")
		}
		if dec.Input == nil {
			return errors.New("missing required field 'input' in transaction")
		}
		inner = &ArbitrumUnsignedTx{
			ChainId:   (*big.Int)(dec.ChainID),
			From:      *dec.From,
			Nonce:     uint64(*dec.Nonce),
			GasFeeCap: (*big.Int)(dec.MaxFeePerGas),
			Gas:       uint64(*dec.Gas),
			To:        dec.To,
			Value:     (*big.Int)(dec.Value),
			Data:      *dec.Input,
		}

	case ArbitrumContractTxType:
		if dec.ChainID == nil {
			return errors.New("missing required field 'chainId' in transaction")
		}
		if dec.RequestId == nil {
			return errors.New("missing required field 'requestId' in transaction")
		}
		if dec.From == nil {
			return errors.New("missing required field 'from' in transaction")
		}
		if dec.MaxFeePerGas == nil {
			return errors.New("missing required field 'maxFeePerGas' for txdata")
		}
		if dec.Gas == nil {
			return errors.New("missing required field 'gas' in txdata")
		}
		if dec.Value == nil {
			return errors.New("missing required field 'value' in transaction")
		}
		if dec.Input == nil {
			return errors.New("missing required field 'input' in transaction")
		}
		inner = &ArbitrumContractTx{
			ChainId:   (*big.Int)(dec.ChainID),
			RequestId: *dec.RequestId,
			From:      *dec.From,
			GasFeeCap: (*big.Int)(dec.MaxFeePerGas),
			Gas:       uint64(*dec.Gas),
			To:        dec.To,
			Value:     (*big.Int)(dec.Value),
			Data:      *dec.Input,
		}

	case ArbitrumRetryTxType:
		if dec.ChainID == nil {
			return errors.New("missing required field 'chainId' in transaction")
		}
		if dec.Nonce == nil {
			return errors.New("missing required field 'nonce' in transaction")
		}
		if dec.From == nil {
			return errors.New("missing required field 'from' in transaction")
		}
		if dec.MaxFeePerGas == nil {
			return errors.New("missing required field 'maxFeePerGas' for txdata")
		}
		if dec.Gas == nil {
			return errors.New("missing required field 'gas' in txdata")
		}
		if dec.Value == nil {
			return errors.New("missing required field 'value' in transaction")
		}
		if dec.Input == nil {
			return errors.New("missing required field 'input' in transaction")
		}
		if dec.TicketId == nil {
			return errors.New("missing required field 'ticketId' in transaction")
		}
		if dec.RefundTo == nil {
			return errors.New("missing required field 'refundTo' in transaction")
		}
		if dec.MaxRefund == nil {
			return errors.New("missing required field 'maxRefund' in transaction")
		}
		if dec.SubmissionFeeRefund == nil {
			return errors.New("missing required field 'submissionFeeRefund' in transaction")
		}
		inner = &ArbitrumRetryTx{
			ChainId:             (*big.Int)(dec.ChainID),
			Nonce:               uint64(*dec.Nonce),
			From:                *dec.From,
			GasFeeCap:           (*big.Int)(dec.MaxFeePerGas),
			Gas:                 uint64(*dec.Gas),
			To:                  dec.To,
			Value:               (*big.Int)(dec.Value),
			Data:                *dec.Input,
			TicketId:            *dec.TicketId,
			RefundTo:            *dec.RefundTo,
			MaxRefund:           (*big.Int)(dec.MaxRefund),
			SubmissionFeeRefund: (*big.Int)(dec.SubmissionFeeRefund),
		}

	case ArbitrumSubmitRetryableTxType:
		if dec.ChainID == nil {
			return errors.New("missing required field 'chainId' in transaction")
		}
		if dec.RequestId == nil {
			return errors.New("missing required field 'requestId' in transaction")
		}
		if dec.From == nil {
			return errors.New("missing required field 'from' in transaction")
		}
		if dec.L1BaseFee == nil {
			return errors.New("missing required field 'l1BaseFee' in transaction")
		}
		if dec.DepositValue == nil {
			return errors.New("missing required field 'depositValue' in transaction")
		}
		if dec.MaxFeePerGas == nil {
			return errors.New("missing required field 'maxFeePerGas' for txdata")
		}
		if dec.Gas == nil {
			return errors.New("missing required field 'gas' in txdata")
		}
		if dec.Beneficiary == nil {
			return errors.New("missing required field 'beneficiary' in transaction")
		}
		if dec.MaxSubmissionFee == nil {
			return errors.New("missing required field 'maxSubmissionFee' in transaction")
		}
		if dec.RefundTo == nil {
			return errors.New("missing required field 'refundTo' in transaction")
		}
		if dec.RetryValue == nil {
			return errors.New("missing required field 'retryValue' in transaction")
		}
		if dec.RetryData == nil {
			return errors.New("missing required field 'retryData' in transaction")
		}
		inner = &ArbitrumSubmitRetryableTx{
			ChainId:          (*big.Int)(dec.ChainID),
			RequestId:        *dec.RequestId,
			From:             *dec.From,
			L1BaseFee:        (*big.Int)(dec.L1BaseFee),
			DepositValue:     (*big.Int)(dec.DepositValue),
			GasFeeCap:        (*big.Int)(dec.MaxFeePerGas),
			Gas:              uint64(*dec.Gas),
			RetryTo:          dec.RetryTo,
			RetryValue:       (*big.Int)(dec.RetryValue),
			Beneficiary:      *dec.Beneficiary,
			MaxSubmissionFee: (*big.Int)(dec.MaxSubmissionFee),
			FeeRefundAddr:    *dec.RefundTo,
			RetryData:        *dec.RetryData,
		}

	default:
		return ErrTxTypeNotSupported
	}

	tx.setDecoded(inner, 0)

	return nil
}
