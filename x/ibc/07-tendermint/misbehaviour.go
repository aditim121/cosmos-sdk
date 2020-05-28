package tendermint

import (
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// CheckMisbehaviourAndUpdateState determines whether or not two conflicting
// headers at the same height would have convinced the light client.
//
// NOTE: assumes provided height is the height at which the consensusState is
// stored.
func CheckMisbehaviourAndUpdateState(
	clientState clientexported.ClientState,
	consensusState clientexported.ConsensusState,
	misbehaviour clientexported.Misbehaviour,
	height uint64, // height at which the consensus state was loaded
	currentTimestamp time.Time,
) (clientexported.ClientState, error) {

	// cast the interface to specific types before checking for misbehaviour
	tmClientState, ok := clientState.(types.ClientState)
	if !ok {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidClientType, "client state type is not Tendermint")
	}

	// If client is already frozen at earlier height than evidence, return with error
	if tmClientState.IsFrozen() && tmClientState.FrozenHeight <= uint64(misbehaviour.GetHeight()) {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence,
			"client is already frozen at earlier height %d than misbehaviour height %d", tmClientState.FrozenHeight, misbehaviour.GetHeight())
	}

	tmConsensusState, ok := consensusState.(types.ConsensusState)
	if !ok {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidClientType, "consensus state is not Tendermint")
	}

	tmEvidence, ok := misbehaviour.(types.Evidence)
	if !ok {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidClientType, "evidence type is not Tendermint")
	}

	if err := checkMisbehaviour(
		tmClientState, tmConsensusState, tmEvidence, height, currentTimestamp,
	); err != nil {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	tmClientState.FrozenHeight = uint64(tmEvidence.GetHeight())
	return tmClientState, nil
}

// checkMisbehaviour checks if the evidence provided is a valid light client misbehaviour
func checkMisbehaviour(
	clientState types.ClientState, consensusState types.ConsensusState, evidence types.Evidence,
	height uint64, currentTimestamp time.Time,
) error {
	// check if provided height matches the headers' height
	if height > uint64(evidence.GetHeight()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"height > evidence header height (%d > %d)", height, evidence.GetHeight(),
		)
	}

	// NOTE: header height and commitment root assertions are checked with the
	// evidence and msg ValidateBasic functions at the AnteHandler level.

	// assert that the timestamp is not from more than an unbonding period ago
	if currentTimestamp.Sub(consensusState.Timestamp) >= clientState.UnbondingPeriod {
		return sdkerrors.Wrapf(
			types.ErrUnbondingPeriodExpired,
			"current timestamp minus the latest consensus state timestamp is greater than or equal to the unbonding period (%s >= %s)",
			currentTimestamp.Sub(consensusState.Timestamp), clientState.UnbondingPeriod,
		)
	}

	valset, err := tmtypes.ValidatorSetFromProto(consensusState.ValidatorSet)
	if err != nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	signedHeader1, err := tmtypes.SignedHeaderFromProto(&evidence.Header1.SignedHeader)
	if err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "invalid signed header 1: %s", err.Error())
	}

	signedHeader2, err := tmtypes.SignedHeaderFromProto(&evidence.Header2.SignedHeader)
	if err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "invalid signed header 2: %s", err.Error())
	}

	// TODO: Evidence must be within trusting period
	// Blocked on https://github.com/cosmos/ics/issues/379

	// - ValidatorSet must have (1-trustLevel) similarity with trusted FromValidatorSet
	// - ValidatorSets on both headers are valid given the last trusted ValidatorSet
	if err := valset.VerifyCommitTrusting(
		evidence.ChainID, signedHeader1.Commit.BlockID, signedHeader1.Height,
		signedHeader1.Commit, clientState.TrustLevel.ToTendermint(),
	); err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "validator set in header 1 has too much change from last known validator set: %v", err)
	}

	if err := valset.VerifyCommitTrusting(
		evidence.ChainID, signedHeader2.Commit.BlockID, signedHeader2.Height,
		signedHeader2.Commit, clientState.TrustLevel.ToTendermint(),
	); err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "validator set in header 2 has too much change from last known validator set: %v", err)
	}

	return nil
}
