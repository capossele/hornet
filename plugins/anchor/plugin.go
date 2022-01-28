package anchor

import (
	"context"
	"encoding/hex"

	"go.uber.org/dig"
	"golang.org/x/crypto/blake2b"

	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/gohornet/hornet/pkg/model/milestone"
	"github.com/gohornet/hornet/pkg/model/storage"
	"github.com/gohornet/hornet/pkg/node"
	"github.com/gohornet/hornet/pkg/proofofinclusion"
	"github.com/gohornet/hornet/pkg/shutdown"
	"github.com/gohornet/hornet/pkg/tangle"
	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/goshimmer/packages/jsonmodels"
	"github.com/iotaledger/hive.go/configuration"
	"github.com/iotaledger/hive.go/events"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/wilfreddenton/merkle"
)

func init() {
	Plugin = &node.Plugin{
		Status: node.StatusDisabled,
		Pluggable: node.Pluggable{
			Name:     "Anchor",
			DepsFunc: func(cDeps dependencies) { deps = cDeps },
			Run:      run,
		},
	}

	h, _ := blake2b.New256(nil)
	h.Write([]byte("Mainnet"))
	ChildTangleID = hex.EncodeToString(h.Sum(nil))
}

var (
	Plugin        *node.Plugin
	deps          dependencies
	c             *client.GoShimmerAPI
	ChildTangleID string
)

type dependencies struct {
	dig.In
	Storage            *storage.Storage
	Tangle             *tangle.Tangle
	NodeConfig         *configuration.Configuration `name:"nodeConfig"`
	RestAPIBindAddress string                       `name:"restAPIBindAddress"`
}

func run() {
	c = client.NewGoShimmerAPI("http://127.0.0.1:8070")

	onConfirmedMilestoneChanged := events.NewClosure(func(cachedMs *storage.CachedMilestone) {
		messagesMemcache := storage.NewMessagesMemcache(deps.Storage)
		metadataMemcache := storage.NewMetadataMemcache(deps.Storage)
		defer func() {
			messagesMemcache.Cleanup(true)
			metadataMemcache.Cleanup(true)
			cachedMs.Release()
		}()

		msIndex := cachedMs.Milestone().Index
		if milestoneMessageID := getMilestoneMessageID(msIndex); milestoneMessageID != nil {
			cachedMsgMeta := deps.Storage.CachedMessageMetadataOrNil(milestoneMessageID)
			if cachedMsgMeta == nil {
				Plugin.LogError("error retrieving cachedMsgMetadata for milestone")
				return
			}
			defer cachedMsgMeta.Release(true)
			parents := cachedMsgMeta.Metadata().Parents()

			ipc, err := proofofinclusion.ComputeIncludedPastCone(context.Background(), deps.Storage, msIndex, metadataMemcache, messagesMemcache, parents)
			if err != nil {
				Plugin.LogErrorf("error computing included past cone %s", err)
				return
			}

			Plugin.LogInfo("Merkle Root -> ", hex.EncodeToString(ipc.MerkleTree.Root()))

			anchorMsg := &jsonmodels.Anchor{
				Version:        1,
				ChildTangleID:  ChildTangleID,
				LastStampID:    ChildTangleID,
				ChildMessageID: milestoneMessageID.ToHex(),
				MerkleRoot:     hex.EncodeToString(ipc.MerkleTree.Root()),
			}
			if err := c.IssueAnchor(anchorMsg); err != nil {
				Plugin.LogErrorf("error issuing anchor %s", err)
			}
		}
	})

	if err := Plugin.Daemon().BackgroundWorker("Anchor", func(ctx context.Context) {
		deps.Tangle.Events.ConfirmedMilestoneChanged.Attach(onConfirmedMilestoneChanged)
		<-ctx.Done()
		Plugin.LogInfo("Stopping Anchoring ...")
		deps.Tangle.Events.ConfirmedMilestoneChanged.Detach(onConfirmedMilestoneChanged)
		Plugin.LogInfo("Stopping Anchoring ... done")
	}, shutdown.PriorityDashboard); err != nil {
		Plugin.LogPanicf("failed to start worker: %s", err)
	}
}

func getMilestoneMessageID(index milestone.Index) hornet.MessageID {
	cachedMs := deps.Storage.MilestoneCachedMessageOrNil(index) // message +1
	if cachedMs == nil {
		return nil
	}
	defer cachedMs.Release(true) // message -1

	return cachedMs.Message().MessageID()
}

func ProofOfInclusion(messageID hornet.MessageID) (*merkle.Tree, hornet.MessageID, error) {
	messagesMemcache := storage.NewMessagesMemcache(deps.Storage)
	metadataMemcache := storage.NewMetadataMemcache(deps.Storage)
	defer func() {
		messagesMemcache.Cleanup(true)
		metadataMemcache.Cleanup(true)
	}()

	cachedMsgMeta := deps.Storage.CachedMessageMetadataOrNil(messageID)
	if cachedMsgMeta == nil {
		return nil, nil, errors.WithMessagef(echo.ErrNotFound, "message not found: %s", messageID.ToHex())
	}
	defer cachedMsgMeta.Release(true)

	metadata := cachedMsgMeta.Metadata()

	referenced, msIndex := metadata.ReferencedWithIndex()
	if !referenced {
		return nil, nil, errors.WithMessagef(echo.ErrNotFound, "message not confirmed: %s", messageID.ToHex())
	}

	if milestoneMessageID := getMilestoneMessageID(msIndex); milestoneMessageID != nil {
		cachedMsgMeta := deps.Storage.CachedMessageMetadataOrNil(milestoneMessageID)
		if cachedMsgMeta == nil {
			return nil, nil, errors.WithMessagef(echo.ErrNotFound, "milestone not found for message: %s", messageID.ToHex())
		}
		defer cachedMsgMeta.Release(true)
		parents := cachedMsgMeta.Metadata().Parents()

		ipc, err := proofofinclusion.ComputeIncludedPastCone(context.Background(), deps.Storage, msIndex, metadataMemcache, messagesMemcache, parents)
		if err != nil {
			return nil, nil, errors.WithMessagef(echo.ErrNotFound, "error computing included past cone: %s", err)
		}

		return ipc.MerkleTree, milestoneMessageID, nil
	}
	return nil, nil, errors.WithMessagef(echo.ErrNotFound, "milestone not found for message: %s", messageID.ToHex())
}
