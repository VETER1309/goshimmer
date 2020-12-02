package marker

import (
	"time"

	"github.com/iotaledger/hive.go/objectstorage"
)

const (
	// PrefixSequence defines the storage prefix for the Sequence.
	PrefixSequence byte = iota

	// PrefixSequenceAliasMapping defines the storage prefix for the SequenceAliasMapping.
	PrefixSequenceAliasMapping
)

// objectStorageOptions contains a list of default settings for the object storage.
var objectStorageOptions = []objectstorage.Option{
	objectstorage.CacheTime(60 * time.Second),
}
