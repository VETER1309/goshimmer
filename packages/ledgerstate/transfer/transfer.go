package transfer

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"
)

type Transfer struct {
	hash    Hash
	inputs  []*OutputReference
	outputs map[address.Address][]*coloredcoins.ColoredBalance
}

func NewTransfer(transferHash Hash) *Transfer {
	return &Transfer{
		hash:    transferHash,
		inputs:  make([]*OutputReference, 0),
		outputs: make(map[address.Address][]*coloredcoins.ColoredBalance),
	}
}

func (transfer *Transfer) GetHash() Hash {
	return transfer.hash
}

func (transfer *Transfer) GetInputs() []*OutputReference {
	return transfer.inputs
}

func (transfer *Transfer) AddInput(input *OutputReference) *Transfer {
	transfer.inputs = append(transfer.inputs, input)

	return transfer
}

func (transfer *Transfer) AddOutput(address address.Address, balance *coloredcoins.ColoredBalance) *Transfer {
	if _, addressExists := transfer.outputs[address]; !addressExists {
		transfer.outputs[address] = make([]*coloredcoins.ColoredBalance, 0)
	}

	transfer.outputs[address] = append(transfer.outputs[address], balance)

	return transfer
}

func (transfer *Transfer) GetOutputs() map[address.Address][]*coloredcoins.ColoredBalance {
	return transfer.outputs
}

func (transfer *Transfer) MarshalBinary() (data []byte, err error) {
	return
}

func (transfer *Transfer) UnmarshalBinary(data []byte) (err error) {
	return
}
