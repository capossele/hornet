package utxo

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gohornet/hornet/pkg/model/milestone"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
)

func TestConfirmationApplyAndRollbackToEmptyLedger(t *testing.T) {

	utxo := New(mapdb.NewMapDB())

	outputs := Outputs{
		randOutput(iotago.OutputExtended),
		randOutput(iotago.OutputExtended),
		randOutput(iotago.OutputNFT),      // spent
		randOutput(iotago.OutputExtended), // spent
		randOutput(iotago.OutputAlias),
		randOutput(iotago.OutputNFT),
		randOutput(iotago.OutputFoundry),
	}

	msIndex := milestone.Index(756)

	spents := Spents{
		randomSpent(outputs[3], msIndex),
		randomSpent(outputs[2], msIndex),
	}

	require.NoError(t, utxo.ApplyConfirmationWithoutLocking(msIndex, outputs, spents, nil, nil))

	var outputCount int
	require.NoError(t, utxo.ForEachOutput(func(_ *Output) bool {
		outputCount++
		return true
	}))
	require.Equal(t, 7, outputCount)

	var unspentExtendedCount int
	require.NoError(t, utxo.ForEachUnspentExtendedOutput(nil, func(_ *Output) bool {
		unspentExtendedCount++
		return true
	}))
	require.Equal(t, 2, unspentExtendedCount)

	var unspentNFTCount int
	require.NoError(t, utxo.ForEachUnspentNFTOutput(nil, func(_ *Output) bool {
		unspentNFTCount++
		return true
	}))
	require.Equal(t, 1, unspentNFTCount)

	var unspentAliasCount int
	require.NoError(t, utxo.ForEachUnspentAliasOutput(nil, func(_ *Output) bool {
		unspentAliasCount++
		return true
	}))
	require.Equal(t, 1, unspentAliasCount)

	var unspentFoundryCount int
	require.NoError(t, utxo.ForEachUnspentFoundryOutput(nil, func(_ *Output) bool {
		unspentFoundryCount++
		return true
	}))
	require.Equal(t, 1, unspentFoundryCount)

	var spentCount int
	require.NoError(t, utxo.ForEachSpentOutput(func(_ *Spent) bool {
		spentCount++
		return true
	}))
	require.Equal(t, 2, spentCount)

	require.NoError(t, utxo.RollbackConfirmationWithoutLocking(msIndex, outputs, spents, nil, nil))

	require.NoError(t, utxo.ForEachOutput(func(_ *Output) bool {
		require.Fail(t, "should not be called")
		return true
	}))

	require.NoError(t, utxo.ForEachUnspentExtendedOutput(nil, func(_ *Output) bool {
		require.Fail(t, "should not be called")
		return true
	}))

	require.NoError(t, utxo.ForEachUnspentNFTOutput(nil, func(_ *Output) bool {
		require.Fail(t, "should not be called")
		return true
	}))

	require.NoError(t, utxo.ForEachUnspentAliasOutput(nil, func(_ *Output) bool {
		require.Fail(t, "should not be called")
		return true
	}))

	require.NoError(t, utxo.ForEachUnspentFoundryOutput(nil, func(_ *Output) bool {
		require.Fail(t, "should not be called")
		return true
	}))

	require.NoError(t, utxo.ForEachSpentOutput(func(_ *Spent) bool {
		require.Fail(t, "should not be called")
		return true
	}))
}

func TestConfirmationApplyAndRollbackToPreviousLedger(t *testing.T) {

	utxo := New(mapdb.NewMapDB())

	previousOutputs := Outputs{
		randOutput(iotago.OutputExtended),
		randOutput(iotago.OutputExtended), // spent
		randOutput(iotago.OutputNFT),      // spent on 2nd confirmation
	}

	previousMsIndex := milestone.Index(48)
	previousSpents := Spents{
		randomSpent(previousOutputs[1], previousMsIndex),
	}
	require.NoError(t, utxo.ApplyConfirmationWithoutLocking(previousMsIndex, previousOutputs, previousSpents, nil, nil))

	ledgerIndex, err := utxo.ReadLedgerIndex()
	require.NoError(t, err)
	require.Equal(t, previousMsIndex, ledgerIndex)

	outputs := Outputs{
		randOutput(iotago.OutputExtended),
		randOutput(iotago.OutputFoundry),
		randOutput(iotago.OutputExtended), // spent
		randOutput(iotago.OutputAlias),
	}
	msIndex := milestone.Index(49)
	spents := Spents{
		randomSpent(previousOutputs[2], msIndex),
		randomSpent(outputs[2], msIndex),
	}
	require.NoError(t, utxo.ApplyConfirmationWithoutLocking(msIndex, outputs, spents, nil, nil))

	ledgerIndex, err = utxo.ReadLedgerIndex()
	require.NoError(t, err)
	require.Equal(t, msIndex, ledgerIndex)

	// Prepare values to check
	outputByOutputID := make(map[string]struct{})
	unspentByOutputID := make(map[string]struct{})
	for _, output := range previousOutputs {
		outputByOutputID[output.mapKey()] = struct{}{}
		unspentByOutputID[output.mapKey()] = struct{}{}
	}
	for _, output := range outputs {
		outputByOutputID[output.mapKey()] = struct{}{}
		unspentByOutputID[output.mapKey()] = struct{}{}
	}

	spentByOutputID := make(map[string]struct{})
	for _, spent := range previousSpents {
		spentByOutputID[spent.mapKey()] = struct{}{}
		delete(unspentByOutputID, spent.mapKey())
	}
	for _, spent := range spents {
		spentByOutputID[spent.mapKey()] = struct{}{}
		delete(unspentByOutputID, spent.mapKey())
	}

	var outputCount int
	require.NoError(t, utxo.ForEachOutput(func(output *Output) bool {
		outputCount++
		_, has := outputByOutputID[output.mapKey()]
		require.True(t, has)
		delete(outputByOutputID, output.mapKey())
		return true
	}))
	require.Empty(t, outputByOutputID)
	require.Equal(t, 7, outputCount)

	var unspentCount int
	require.NoError(t, utxo.ForEachUnspentExtendedOutput(nil, func(output *Output) bool {
		unspentCount++
		_, has := unspentByOutputID[output.mapKey()]
		require.True(t, has)
		delete(unspentByOutputID, output.mapKey())
		return true
	}))
	require.Equal(t, 2, unspentCount)
	require.NoError(t, utxo.ForEachUnspentNFTOutput(nil, func(output *Output) bool {
		unspentCount++
		_, has := unspentByOutputID[output.mapKey()]
		require.True(t, has)
		delete(unspentByOutputID, output.mapKey())
		return true
	}))
	require.Equal(t, 2, unspentCount)
	require.NoError(t, utxo.ForEachUnspentAliasOutput(nil, func(output *Output) bool {
		unspentCount++
		_, has := unspentByOutputID[output.mapKey()]
		require.True(t, has)
		delete(unspentByOutputID, output.mapKey())
		return true
	}))
	require.Equal(t, 3, unspentCount)
	require.NoError(t, utxo.ForEachUnspentFoundryOutput(nil, func(output *Output) bool {
		unspentCount++
		_, has := unspentByOutputID[output.mapKey()]
		require.True(t, has)
		delete(unspentByOutputID, output.mapKey())
		return true
	}))
	require.Equal(t, 4, unspentCount)
	require.Empty(t, unspentByOutputID)

	var spentCount int
	require.NoError(t, utxo.ForEachSpentOutput(func(spent *Spent) bool {
		spentCount++
		_, has := spentByOutputID[spent.mapKey()]
		require.True(t, has)
		delete(spentByOutputID, spent.mapKey())
		return true
	}))
	require.Empty(t, spentByOutputID)
	require.Equal(t, 3, spentCount)

	require.NoError(t, utxo.RollbackConfirmationWithoutLocking(msIndex, outputs, spents, nil, nil))

	ledgerIndex, err = utxo.ReadLedgerIndex()
	require.NoError(t, err)
	require.Equal(t, previousMsIndex, ledgerIndex)

	// Prepare values to check
	outputByOutputID = make(map[string]struct{})
	unspentByOutputID = make(map[string]struct{})
	spentByOutputID = make(map[string]struct{})

	for _, output := range previousOutputs {
		outputByOutputID[output.mapKey()] = struct{}{}
		unspentByOutputID[output.mapKey()] = struct{}{}
	}

	for _, spent := range previousSpents {
		spentByOutputID[spent.mapKey()] = struct{}{}
		delete(unspentByOutputID, spent.mapKey())
	}

	require.NoError(t, utxo.ForEachOutput(func(output *Output) bool {
		_, has := outputByOutputID[output.mapKey()]
		require.True(t, has)
		delete(outputByOutputID, output.mapKey())
		return true
	}))
	require.Empty(t, outputByOutputID)

	require.NoError(t, utxo.ForEachUnspentExtendedOutput(nil, func(output *Output) bool {
		_, has := unspentByOutputID[output.mapKey()]
		require.True(t, has)
		delete(unspentByOutputID, output.mapKey())
		return true
	}))
	require.NoError(t, utxo.ForEachUnspentNFTOutput(nil, func(output *Output) bool {
		_, has := unspentByOutputID[output.mapKey()]
		require.True(t, has)
		delete(unspentByOutputID, output.mapKey())
		return true
	}))
	require.NoError(t, utxo.ForEachUnspentAliasOutput(nil, func(output *Output) bool {
		_, has := unspentByOutputID[output.mapKey()]
		require.True(t, has)
		delete(unspentByOutputID, output.mapKey())
		return true
	}))
	require.NoError(t, utxo.ForEachUnspentFoundryOutput(nil, func(output *Output) bool {
		_, has := unspentByOutputID[output.mapKey()]
		require.True(t, has)
		delete(unspentByOutputID, output.mapKey())
		return true
	}))
	require.Empty(t, unspentByOutputID)

	require.NoError(t, utxo.ForEachSpentOutput(func(spent *Spent) bool {
		_, has := spentByOutputID[spent.mapKey()]
		require.True(t, has)
		delete(spentByOutputID, spent.mapKey())
		return true
	}))
	require.Empty(t, spentByOutputID)
}
