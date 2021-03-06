package payload

import (
	"fmt"
	"testing"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/clock"
	"github.com/iotaledger/goshimmer/packages/tangle"
)

func ExamplePayload() {
	// 1. create value transfer (user provides this)
	valueTransfer := transaction.New(
		// inputs
		transaction.NewInputs(
			transaction.NewOutputID(address.Random(), transaction.RandomID()),
			transaction.NewOutputID(address.Random(), transaction.RandomID()),
		),

		// outputs
		transaction.NewOutputs(map[address.Address][]*balance.Balance{
			address.Random(): {
				balance.New(balance.ColorIOTA, 1337),
			},
		}),
	)

	// 2. create value payload (the ontology creates this and wraps the user provided transfer accordingly)
	valuePayload := New(
		// parent1 in "value transfer ontology" (filled by ontology tipSelector)
		GenesisID,

		// parent2 in "value transfer ontology"  (filled by ontology tipSelector)
		GenesisID,

		// value transfer
		valueTransfer,
	)

	// 3. build actual transaction (the base layer creates this and wraps the ontology provided payload)
	tx := tangle.NewMessage(
		[]tangle.MessageID{tangle.EmptyMessageID},
		[]tangle.MessageID{},

		// the time when the transaction was created
		clock.SyncedTime(),

		// public key of the issuer
		ed25519.PublicKey{},

		// the ever increasing sequence number of this transaction
		0,

		// payload
		valuePayload,

		// nonce to check PoW
		0,

		// signature
		ed25519.Signature{},
	)

	fmt.Println(tx)
}

func TestPayload(t *testing.T) {
	addressKeyPair1 := ed25519.GenerateKeyPair()
	addressKeyPair2 := ed25519.GenerateKeyPair()

	originalPayload := New(
		GenesisID,
		GenesisID,
		transaction.New(
			transaction.NewInputs(
				transaction.NewOutputID(address.FromED25519PubKey(addressKeyPair1.PublicKey), transaction.RandomID()),
				transaction.NewOutputID(address.FromED25519PubKey(addressKeyPair2.PublicKey), transaction.RandomID()),
			),

			transaction.NewOutputs(map[address.Address][]*balance.Balance{
				address.Random(): {
					balance.New(balance.ColorIOTA, 1337),
				},
			}),
		).Sign(
			signaturescheme.ED25519(addressKeyPair1),
		),
	)

	assert.Equal(t, false, originalPayload.Transaction().SignaturesValid())

	originalPayload.Transaction().Sign(
		signaturescheme.ED25519(addressKeyPair2),
	)

	assert.Equal(t, true, originalPayload.Transaction().SignaturesValid())

	clonedPayload1, _, err := FromBytes(originalPayload.Bytes())
	if err != nil {
		panic(err)
	}

	assert.Equal(t, originalPayload.Parent2ID(), clonedPayload1.Parent2ID())
	assert.Equal(t, originalPayload.Parent1ID(), clonedPayload1.Parent1ID())
	assert.Equal(t, originalPayload.Transaction().Bytes(), clonedPayload1.Transaction().Bytes())
	assert.Equal(t, originalPayload.ID(), clonedPayload1.ID())
	assert.Equal(t, true, clonedPayload1.Transaction().SignaturesValid())

	clonedPayload2, _, err := FromBytes(clonedPayload1.Bytes())
	if err != nil {
		panic(err)
	}

	assert.Equal(t, originalPayload.ID(), clonedPayload2.ID())
	assert.Equal(t, true, clonedPayload2.Transaction().SignaturesValid())
}
