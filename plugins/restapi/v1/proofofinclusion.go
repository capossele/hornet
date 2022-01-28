package v1

import (
	"encoding/hex"

	"github.com/gohornet/hornet/pkg/proofofinclusion"
	"github.com/gohornet/hornet/pkg/restapi"
	"github.com/gohornet/hornet/plugins/anchor"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/wilfreddenton/merkle"
	"golang.org/x/crypto/blake2b"
)

func computeProofOfInclusion(c echo.Context) (*proofOfInclusionResponse, error) {

	if !deps.SyncManager.IsNodeAlmostSynced() {
		return nil, errors.WithMessage(echo.ErrServiceUnavailable, "node is not synced")
	}

	messageID, err := restapi.ParseMessageIDParam(c)
	if err != nil {
		return nil, err
	}

	tree, milestoneMessageID, err := anchor.ProofOfInclusion(messageID)
	if err != nil {
		return nil, err
	}

	h, _ := blake2b.New256(nil)
	target := merkle.LeafHash(messageID, h)
	path := tree.MerklePath(target)

	response := &proofOfInclusionResponse{
		MessageID:   messageID.ToHex(),
		MilestoneID: milestoneMessageID.ToHex(),
		MerkleRoot:  hex.EncodeToString(tree.Root()),
		Path:        proofofinclusion.PathToStrings(path),
	}

	return response, nil
}
