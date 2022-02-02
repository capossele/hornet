package proofofinclusion

import (
	"context"
	"encoding/hex"

	"github.com/gohornet/hornet/pkg/dag"
	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/gohornet/hornet/pkg/model/milestone"
	"github.com/gohornet/hornet/pkg/model/storage"

	"github.com/wilfreddenton/merkle"
	"golang.org/x/crypto/blake2b"
)

// IncludedPastCone contains the ledger mutations and referenced messages applied to a cone under the "white-flag" approach.
type IncludedPastCone struct {
	// The messages in the order in which they were applied.
	MessagesIncluded hornet.MessageIDs
	// The merkle tree of all messages.
	MerkleTree *merkle.Tree
}

func ComputeIncludedPastCone(ctx context.Context, dbStorage *storage.Storage, msIndex milestone.Index, metadataMemcache *storage.MetadataMemcache, messagesMemcache *storage.MessagesMemcache, parents hornet.MessageIDs) (*IncludedPastCone, error) {
	ipc := &IncludedPastCone{
		MessagesIncluded: make(hornet.MessageIDs, 0),
	}

	// traversal stops if no more messages pass the given condition
	// Caution: condition func is not in DFS order
	condition := func(cachedMetadata *storage.CachedMetadata) (bool, error) { // meta +1
		defer cachedMetadata.Release(true) // meta -1

		referenced, index := cachedMetadata.Metadata().ReferencedWithIndex()
		return referenced && index == msIndex, nil
		// only traverse and process the message if it was not referenced yet
		// return !cachedMetadata.Metadata().IsReferenced(), nil
	}

	// consumer
	consumer := func(cachedMetadata *storage.CachedMetadata) error { // meta +1
		defer cachedMetadata.Release(true) // meta -1
		ipc.MessagesIncluded = append(ipc.MessagesIncluded, cachedMetadata.Metadata().MessageID())
		return nil
	}

	// we don't need to call cleanup at the end, because we pass our own metadataMemcache.
	parentsTraverser := dag.NewParentTraverser(dbStorage, metadataMemcache)

	// This function does the DFS and computes the mutations a white-flag confirmation would create.
	// If the parents are SEPs, are already processed or already referenced,
	// then the mutations from the messages retrieved from the stack are accumulated to the given Confirmation struct's mutations.
	// If the popped message was used to mutate the Confirmation struct, it will also be appended to Confirmation.MessagesIncludedWithTransactions.
	if err := parentsTraverser.Traverse(
		ctx,
		parents,
		condition,
		consumer,
		// called on missing parents
		// return error on missing parents
		nil,
		// called on solid entry points
		// Ignore solid entry points (snapshot milestone included)
		nil,
		false); err != nil {
		return nil, err
	}

	// compute merkle tree root hash
	marshalers := make([][]byte, len(ipc.MessagesIncluded))
	for i := range ipc.MessagesIncluded {
		marshalers[i] = ipc.MessagesIncluded[i]
	}

	h, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	// initialize the tree
	t := merkle.NewTree()
	// compute the root hash from the pre-leaves using the sha256 hash function
	err = t.Hash(marshalers, h)
	if err != nil {
		return nil, err
	}

	ipc.MerkleTree = t

	return ipc, nil
}

func PathToStrings(path []*merkle.Node) []string {
	output := make([]string, len(path))
	for i, p := range path {
		output[i] = hex.EncodeToString(p.Hash)
	}
	return output
}
