package ledgerstate

import (
	"crypto/rand"
	"fmt"
	"strconv"

	"github.com/iotaledger/goshimmer/packages/cerrors"
	"github.com/iotaledger/goshimmer/packages/tangle/payload"
	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/stringify"
	"github.com/iotaledger/hive.go/typeutils"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// region TransactionType //////////////////////////////////////////////////////////////////////////////////////////////

// TransactionType represents the payload Type of a Transaction.
var TransactionType payload.Type

// init defers the initialization of the TransactionType to not have an initialization loop.
func init() {
	TransactionType = payload.NewType(1, "TransactionType", func(data []byte) (payload.Payload, error) {
		return TransactionFromMarshalUtil(marshalutil.New(data))
	})
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region TransactionID ////////////////////////////////////////////////////////////////////////////////////////////////

// TransactionIDLength contains the amount of bytes that a marshaled version of the ID contains.
const TransactionIDLength = 32

// TransactionID is the type that represents the identifier of a Transaction.
type TransactionID [TransactionIDLength]byte

// GenesisTransactionID represents the identifier of the genesis Transaction.
var GenesisTransactionID TransactionID

// TransactionIDFromBytes unmarshals a TransactionID from a sequence of bytes.
func TransactionIDFromBytes(bytes []byte) (transactionID TransactionID, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if transactionID, err = TransactionIDFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse TransactionID from MarshalUtil: %w", err)
		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// TransactionIDFromBase58 creates a TransactionID from a base58 encoded string.
func TransactionIDFromBase58(base58String string) (transactionID TransactionID, err error) {
	bytes, err := base58.Decode(base58String)
	if err != nil {
		err = xerrors.Errorf("error while decoding base58 encoded TransactionID (%v): %w", err, cerrors.Base58DecodeFailed)
		return
	}

	if transactionID, _, err = TransactionIDFromBytes(bytes); err != nil {
		err = xerrors.Errorf("failed to parse TransactionID from bytes: %w", err)
		return
	}

	return
}

// TransactionIDFromMarshalUtil unmarshals a TransactionID using a MarshalUtil (for easier unmarshaling).
func TransactionIDFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (transactionID TransactionID, err error) {
	transactionIDBytes, err := marshalUtil.ReadBytes(TransactionIDLength)
	if err != nil {
		err = xerrors.Errorf("failed to parse TransactionID (%v): %w", err, cerrors.ParseBytesFailed)
		return
	}
	copy(transactionID[:], transactionIDBytes)

	return
}

// TransactionIDFromRandomness returns a random TransactionID which can for example be used in unit tests.
func TransactionIDFromRandomness() (transactionID TransactionID, err error) {
	_, err = rand.Read(transactionID[:])

	return
}

// Bytes marshals the ID into a sequence of bytes.
func (i TransactionID) Bytes() []byte {
	return i[:]
}

// Base58 returns a base58 encoded version of the TransactionID.
func (i TransactionID) Base58() string {
	return base58.Encode(i[:])
}

// String creates a human readable base58 encoded version of the TransactionID.
func (i TransactionID) String() string {
	return "TransactionID(" + i.Base58() + ")"
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Transaction //////////////////////////////////////////////////////////////////////////////////////////////////

// Transaction represents a payload that is executing a value transfer in the ledger state.
type Transaction struct {
	essence      *TransactionEssence
	unlockBlocks UnlockBlocks
}

// NewTransaction create a new Transaction from the given details.
func NewTransaction(essence *TransactionEssence, unlockBlocks UnlockBlocks) *Transaction {
	if len(unlockBlocks) != len(essence.Inputs()) {
		panic(fmt.Sprintf("amount of UnlockBlocks (%d) does not match amount of Inputs (%d)", len(unlockBlocks), len(essence.inputs)))
	}

	return &Transaction{
		essence:      essence,
		unlockBlocks: unlockBlocks,
	}
}

// TransactionFromBytes unmarshals an Transaction from a sequence of bytes.
func TransactionFromBytes(bytes []byte) (transaction *Transaction, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if transaction, err = TransactionFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Transaction from MarshalUtil: %w", err)
		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// TransactionFromMarshalUtil unmarshals a Transaction using a MarshalUtil (for easier unmarshaling).
func TransactionFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (transaction *Transaction, err error) {
	readStartOffset := marshalUtil.ReadOffset()

	payloadSize, err := marshalUtil.ReadUint32()
	if err != nil {
		err = xerrors.Errorf("failed to parse payload size from MarshalUtil (%v): %w", err, cerrors.ParseBytesFailed)
		return
	}
	payloadType, err := payload.TypeFromMarshalUtil(marshalUtil)
	if err != nil {
		err = xerrors.Errorf("failed to parse payload Type from MarshalUtil: %w", err)
		return
	}
	if payloadType != TransactionType {
		err = xerrors.Errorf("payload type '%s' does not match expected '%s': %w", payloadType, TransactionType, cerrors.ParseBytesFailed)
		return
	}

	transaction = &Transaction{}
	if transaction.essence, err = TransactionEssenceFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse TransactionEssence from MarshalUtil: %w", err)
		return
	}
	if transaction.unlockBlocks, err = UnlockBlocksFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse UnlockBlocks from MarshalUtil: %w", err)
		return
	}

	parsedBytes := marshalUtil.ReadOffset() - readStartOffset
	if parsedBytes != int(payloadSize)+4 {
		err = xerrors.Errorf("parsed bytes (%d) did not match expected size (%d): %w", parsedBytes, payloadSize, cerrors.ParseBytesFailed)
		return
	}

	if len(transaction.unlockBlocks) != len(transaction.essence.Inputs()) {
		err = xerrors.Errorf("amount of UnlockBlocks (%d) does not match amount of Inputs (%d): %w", len(transaction.unlockBlocks), len(transaction.essence.inputs), cerrors.ParseBytesFailed)
		return
	}

	return
}

// Type returns the Type of the Payload.
func (t *Transaction) Type() payload.Type {
	return TransactionType
}

func (t *Transaction) Essence() *TransactionEssence {
	return t.essence
}

func (t *Transaction) UnlockBlocks() UnlockBlocks {
	return t.unlockBlocks
}

// Bytes returns a marshaled version of the Transaction.
func (t *Transaction) Bytes() []byte {
	if t == nil {
		return marshalutil.New(marshalutil.UINT16_SIZE).WriteUint16(0).Bytes()
	}

	payloadBytes := byteutils.ConcatBytes(TransactionType.Bytes(), t.essence.Bytes(), t.unlockBlocks.Bytes())
	payloadBytesLength := len(payloadBytes)

	return marshalutil.New(marshalutil.UINT16_SIZE + payloadBytesLength).
		WriteUint16(uint16(payloadBytesLength)).
		WriteBytes(payloadBytes).
		Bytes()
}

// String returns a human readable version of the Transaction.
func (t *Transaction) String() string {
	return stringify.Struct("Transaction",
		stringify.StructField("essence", t.Essence()),
		stringify.StructField("unlockBlocks", t.UnlockBlocks()),
	)
}

// code contract (make sure the struct implements all required methods)
var _ payload.Payload = &Transaction{}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region TransactionEssence ///////////////////////////////////////////////////////////////////////////////////////////

// TransactionEssence contains the relevant information of the transfer (without the unlocking information).
type TransactionEssence struct {
	version TransactionEssenceVersion
	inputs  Inputs
	outputs Outputs
	payload payload.Payload
}

func NewTransactionEssence(version TransactionEssenceVersion, inputs Inputs, outputs Outputs) *TransactionEssence {
	return &TransactionEssence{
		version: version,
		inputs:  inputs,
		outputs: outputs,
	}
}

// TransactionEssenceFromBytes unmarshals an Transaction from a sequence of bytes.
func TransactionEssenceFromBytes(bytes []byte) (transactionEssence *TransactionEssence, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if transactionEssence, err = TransactionEssenceFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse TransactionEssence from MarshalUtil: %w", err)
		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// TransactionEssenceFromMarshalUtil unmarshals a TransactionEssence using a MarshalUtil (for easier unmarshaling).
func TransactionEssenceFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (transactionEssence *TransactionEssence, err error) {
	transactionEssence = &TransactionEssence{}
	if transactionEssence.version, err = TransactionEssenceVersionFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse TransactionEssenceVersion from MarshalUtil: %w", err)
		return
	}
	if transactionEssence.inputs, err = InputsFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Inputs from MarshalUtil: %w", err)
		return
	}
	if transactionEssence.outputs, err = OutputsFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Outputs from MarshalUtil: %w", err)
		return
	}
	if transactionEssence.payload, err = payload.FromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Payload from MarshalUtil: %w", err)
		return
	}

	return
}

// Inputs returns the Inputs of the TransactionEssence.
func (t *TransactionEssence) Inputs() Inputs {
	return t.inputs
}

func (t *TransactionEssence) Outputs() Outputs {
	return t.outputs
}

// Bytes returns a marshaled version of the TransactionEssence.
func (t *TransactionEssence) Bytes() []byte {
	marshalUtil := marshalutil.New().
		Write(t.version).
		Write(t.inputs).
		Write(t.outputs)

	if !typeutils.IsInterfaceNil(t.payload) {
		marshalUtil.Write(t.payload)
	} else {
		marshalUtil.WriteUint32(0)
	}

	return marshalUtil.Bytes()
}

func (t *TransactionEssence) String() string {
	return stringify.Struct("TransactionEssence",
		stringify.StructField("version", t.version),
		stringify.StructField("inputs", t.inputs),
		stringify.StructField("outputs", t.outputs),
		stringify.StructField("payload", t.payload),
	)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region TransactionEssenceVersion ////////////////////////////////////////////////////////////////////////////////////

// TransactionEssenceVersion represents a byte denoting a version augmented with some additional logic.
type TransactionEssenceVersion uint8

// TransactionEssenceVersionFromBytes unmarshals a TransactionEssenceVersion from a sequence of bytes.
func TransactionEssenceVersionFromBytes(bytes []byte) (version TransactionEssenceVersion, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if version, err = TransactionEssenceVersionFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse version TransactionEssenceVersion from MarshalUtil: %w", err)
		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// TransactionEssenceVersionFromMarshalUtil unmarshals a TransactionEssenceVersion using a MarshalUtil (for easier
// unmarshaling).
func TransactionEssenceVersionFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (version TransactionEssenceVersion, err error) {
	readByte, err := marshalUtil.ReadByte()
	if err != nil {
		err = xerrors.Errorf("failed to parse version TransactionEssenceVersion: %w", err)
		return
	}
	if readByte != 0 {
		err = xerrors.Errorf("failed to parse version TransactionEssenceVersion: %w", err)
		return
	}
	version = TransactionEssenceVersion(readByte)

	return
}

// Bytes returns a marshaled version of the TransactionEssenceVersion.
func (v TransactionEssenceVersion) Bytes() []byte {
	return []byte{byte(v)}
}

// Compare offers a comparator for TransactionEssenceVersions which returns -1 if otherInput is bigger, 1 if it is
// smaller and 0 if they are the same.
func (v TransactionEssenceVersion) Compare(other TransactionEssenceVersion) int {
	switch {
	case v < other:
		return -1
	case v > other:
		return 1
	default:
		return 0
	}
}

// String returns a human readable version of the TransactionEssenceVersion.
func (v TransactionEssenceVersion) String() string {
	return "TransactionEssenceVersion(" + strconv.Itoa(int(v)) + ")"
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
