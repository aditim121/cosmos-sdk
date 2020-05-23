package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

var (
	_ exported.Proof  = (*SignatureProof)(nil)
	_ exported.Prefix = (*SignaturePrefix)(nil)
)

// NewSignaturePrefix constructs a new SignaturePrefix instance.
func NewSignaturePrefix(keyPrefix []byte) SignaturePrefix {
	return SignaturePrefix{
		KeyPrefix: keyPrefix,
	}
}

// GetCommitmentType implements Prefix interface.
func (sp SignaturePrefix) GetCommitmentType() exported.Type {
	return exported.Signature
}

// Bytes returns the key prefix bytes.
func (sp SignaturePrefix) Bytes() []byte {
	return sp.KeyPrefix
}

// IsEmpty returns true if the prefix is empty.
func (sp SignaturePrefix) IsEmpty() bool {
	return len(sp.Bytes()) == 0
}

// ToAny converts signature prefix to protobuf Any.
func (sp SignaturePrefix) ToAny() (*cdctypes.Any, error) {
	any, err := cdctypes.NewAnyWithValue(&sp)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrProtobufAny, "prefix %T: %s", sp, err.Error())
	}
	return any, nil
}

// GetCommitmentType implements ProofI.
func (SignatureProof) GetCommitmentType() exported.Type {
	return exported.Signature
}

// VerifyMembership implements ProofI.
func (SignatureProof) VerifyMembership(exported.Root, exported.Path, []byte) error {
	return nil
}

// VerifyNonMembership implements ProofI.
func (SignatureProof) VerifyNonMembership(exported.Root, exported.Path) error {
	return nil
}

// IsEmpty returns true if the signature is empty.
func (proof SignatureProof) IsEmpty() bool {
	return len(proof.Signature) == 0
}

// ToAny converts signature proof to protobuf Any.
func (proof SignatureProof) ToAny() (*cdctypes.Any, error) {
	any, err := cdctypes.NewAnyWithValue(&proof)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrProtobufAny, "proof %T: %s", proof, err.Error())
	}
	return any, nil
}

// ValidateBasic checks if the proof is empty.
func (proof SignatureProof) ValidateBasic() error {
	if proof.IsEmpty() {
		return ErrInvalidProof
	}
	return nil
}
