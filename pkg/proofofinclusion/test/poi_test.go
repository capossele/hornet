package test

import (
	"encoding/hex"
	"fmt"
	"math"
	"testing"

	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/gohornet/hornet/pkg/testsuite"
	"github.com/gohornet/hornet/pkg/testsuite/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wilfreddenton/merkle"
	"golang.org/x/crypto/blake2b"
)

var (
	seed1, _ = hex.DecodeString("96d9ff7a79e4b0a5f3e5848ae7867064402da92a62eabb4ebbe463f12d1f3b1aace1775488f51cb1e3a80732a03ef60b111d6833ab605aa9f8faebeb33bbe3d9")

	showConfirmationGraphs = false
	MinPoWScore            = 100.0
	BelowMaxDepth          = 15
)

func TestIncludedPastCone(t *testing.T) {

	messageIDsMap := make(map[string]string)

	genesisWallet := utils.NewHDWallet("Seed1", seed1, 0)
	genesisAddress := genesisWallet.Address()

	te := testsuite.SetupTestEnvironment(t, genesisAddress, 3, BelowMaxDepth, MinPoWScore, showConfirmationGraphs)
	defer te.CleanupTestEnvironment(!showConfirmationGraphs)

	//Add token supply to our local HDWallet
	genesisWallet.BookOutput(te.GenesisOutput)

	// Issue some transactions
	messageA := te.NewMessageBuilder("A").Parents(hornet.MessageIDs{te.Milestones[0].Milestone().MessageID, te.Milestones[1].Milestone().MessageID}).BuildIndexation().Store()
	messageIDsMap[messageA.StoredMessageID().ToHex()] = "msgA"
	messageB := te.NewMessageBuilder("B").Parents(hornet.MessageIDs{messageA.StoredMessageID(), te.Milestones[0].Milestone().MessageID}).BuildIndexation().Store()
	messageIDsMap[messageB.StoredMessageID().ToHex()] = "msgB"
	messageC := te.NewMessageBuilder("C").Parents(hornet.MessageIDs{te.Milestones[2].Milestone().MessageID, te.Milestones[0].Milestone().MessageID}).BuildIndexation().Store()
	messageIDsMap[messageC.StoredMessageID().ToHex()] = "msgC"
	messageD := te.NewMessageBuilder("D").Parents(hornet.MessageIDs{messageB.StoredMessageID(), messageC.StoredMessageID()}).BuildIndexation().Store()
	messageIDsMap[messageD.StoredMessageID().ToHex()] = "msgD"
	messageE := te.NewMessageBuilder("E").Parents(hornet.MessageIDs{messageB.StoredMessageID(), messageA.StoredMessageID()}).BuildIndexation().Store()
	messageIDsMap[messageE.StoredMessageID().ToHex()] = "msgE"

	// Confirming milestone include all msg up to message E. This should only include A, B and E
	wfc, confStats := te.IssueAndConfirmMilestoneOnTips(hornet.MessageIDs{messageE.StoredMessageID()}, true)
	require.Equal(t, 3+1, confStats.MessagesReferenced) // A, B, E + 1 for Milestone
	require.Equal(t, 0, confStats.MessagesIncludedWithTransactions)
	require.Equal(t, 0, confStats.MessagesExcludedWithConflictingTransactions)
	require.Equal(t, 3+1, confStats.MessagesExcludedWithoutTransactions) // 1 is for the milestone itself

	cachedMsgMeta := te.Storage().CachedMessageMetadataOrNil(wfc.MilestoneMessageID)
	require.NotNil(t, cachedMsgMeta)
	defer cachedMsgMeta.Release(true)

	fmt.Println("--------->", cachedMsgMeta.Metadata().Parents())
	res, err := te.ComputeIncludedPastCone(cachedMsgMeta.Metadata().Parents(), wfc.MilestoneIndex)
	require.NoError(t, err)

	fmt.Println("--------->", res)

	for _, m := range res.MessagesIncluded {
		fmt.Println(messageIDsMap[m.ToHex()])
	}

	fmt.Println(res.MerkleTree.Root())

	h, _ := blake2b.New256(nil)
	target := merkle.LeafHash(messageA.StoredMessageID(), h)

	path := res.MerkleTree.MerklePath(target)

	for _, p := range path {
		fmt.Println(p.Hash)
	}

	assert.True(t, merkle.Prove(target, res.MerkleTree.Root(), path, h))

	// Issue another message
	messageF := te.NewMessageBuilder("F").Parents(hornet.MessageIDs{messageD.StoredMessageID(), messageE.StoredMessageID()}).BuildIndexation().Store()

	// Confirming milestone at message F. This should confirm D, C and F
	wfc, confStats = te.IssueAndConfirmMilestoneOnTips(hornet.MessageIDs{messageF.StoredMessageID()}, true)

	require.Equal(t, 3+1, confStats.MessagesReferenced) // D, C, F + 1 for Milestone
	require.Equal(t, 0, confStats.MessagesIncludedWithTransactions)
	require.Equal(t, 0, confStats.MessagesExcludedWithConflictingTransactions)
	require.Equal(t, 3+1, confStats.MessagesExcludedWithoutTransactions) // 1 is for the milestone itself

	cachedMsgMeta2 := te.Storage().CachedMessageMetadataOrNil(wfc.MilestoneMessageID)
	require.NotNil(t, cachedMsgMeta2)
	defer cachedMsgMeta2.Release(true)

	fmt.Println("--------->", cachedMsgMeta2.Metadata().Parents())
	res, err = te.ComputeIncludedPastCone(cachedMsgMeta2.Metadata().Parents(), wfc.MilestoneIndex)
	require.NoError(t, err)

	fmt.Println("--------->", res)

	for _, m := range res.MessagesIncluded {
		fmt.Println(messageIDsMap[m.ToHex()])
	}

	fmt.Println(res.MerkleTree.Root())

	target = merkle.LeafHash(messageF.StoredMessageID(), h)

	path = res.MerkleTree.MerklePath(target)

	for _, p := range path {
		fmt.Println(p.Hash)
	}

	assert.True(t, merkle.Prove(target, res.MerkleTree.Root(), path, h))
}

func TestProof(t *testing.T) {
	input := []string{"A", "B", "C", "D", "E"}
	target := "A"
	height := int(math.Log2(float64(len(input))))
	fmt.Println("height", height)
	expected := []string{"A", "B", "AB", "CD"}

	output := make([]string, 0)
	output = append(output, target)

	// for level:=0; level<height; level++{
	// 	node[j] := input[i]+input[i+1]
	// }

	assert.Equal(t, expected, output)
}
