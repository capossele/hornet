package testsuite

import (
	"context"

	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/gohornet/hornet/pkg/model/milestone"
	"github.com/gohornet/hornet/pkg/model/storage"
	"github.com/gohornet/hornet/pkg/proofofinclusion"
)

func (te *TestEnvironment) ComputeIncludedPastCone(entrypoint hornet.MessageIDs, milestoneIndex milestone.Index) (*proofofinclusion.IncludedPastCone, error) {

	messagesMemcache := storage.NewMessagesMemcache(te.storage)
	metadataMemcache := storage.NewMetadataMemcache(te.storage)

	defer func() {
		// all releases are forced since the cone is referenced and not needed anymore

		// release all messages at the end
		messagesMemcache.Cleanup(true)

		// Release all message metadata at the end
		metadataMemcache.Cleanup(true)
	}()

	return proofofinclusion.ComputeIncludedPastCone(context.Background(), te.storage, milestoneIndex, metadataMemcache, messagesMemcache, entrypoint)
}
