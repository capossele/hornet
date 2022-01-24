package anchor

import (
	"context"

	"go.uber.org/dig"

	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/gohornet/hornet/pkg/model/milestone"
	"github.com/gohornet/hornet/pkg/model/storage"
	"github.com/gohornet/hornet/pkg/node"
	"github.com/gohornet/hornet/pkg/shutdown"
	"github.com/gohornet/hornet/pkg/tangle"
	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/goshimmer/packages/jsonmodels"
	"github.com/iotaledger/hive.go/configuration"
	"github.com/iotaledger/hive.go/events"
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
}

var (
	Plugin *node.Plugin
	deps   dependencies
	c      *client.GoShimmerAPI
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

	onLatestMilestoneIndexChanged := events.NewClosure(func(msIndex milestone.Index) {
		if milestoneMessageID := getMilestoneMessageID(msIndex); milestoneMessageID != nil {
			// hub.BroadcastMsg(&Msg{Type: MsgTypeMs, Data: &LivefeedMilestone{MessageID: milestoneMessageID.ToHex(), Index: msIndex}})
			anchorMsg := &jsonmodels.ChatRequest{
				From:    "Chrysalis-Tangle",
				To:      "Devnet-Tangle",
				Message: milestoneMessageID.ToHex(),
			}
			if err := c.SendChatMessage(anchorMsg); err != nil {
				Plugin.LogErrorf("Error issuing anchor %s", err)
			}
		}
	})

	if err := Plugin.Daemon().BackgroundWorker("Anchor", func(ctx context.Context) {
		deps.Tangle.Events.LatestMilestoneIndexChanged.Attach(onLatestMilestoneIndexChanged)
		<-ctx.Done()
		Plugin.LogInfo("Stopping Anchoring ...")
		deps.Tangle.Events.LatestMilestoneIndexChanged.Detach(onLatestMilestoneIndexChanged)
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
