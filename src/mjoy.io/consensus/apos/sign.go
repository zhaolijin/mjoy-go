////////////////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 The mjoy-go Authors.
//
// The mjoy-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// @File: apos_signing.go
// @Date: 2018/06/13 11:12:13
////////////////////////////////////////////////////////////////////////////////

package apos

import (
	"math/big"
	"fmt"
	"errors"
	"mjoy.io/common/types"
	"crypto/ecdsa"
	"mjoy.io/common"
	"mjoy.io/utils/crypto"
)

//go:generate msgp

var (
	ErrInvalidSig = errors.New("invalid  v, r, s values")
	ErrInvalidChainId = errors.New("invalid chain id for signer")
)

// Signer encapsulates apos signature handling. Note that this interface is not a
// stable API and may change at any time to accommodate new protocol rules.
type signer interface {
	// sign the obj
	sign(prv *ecdsa.PrivateKey) (R *big.Int, S *big.Int, V *big.Int, err error)

	// Sender returns the sender address of the Credential.
	sender() (types.Address, error)

	// hash
	hash() types.Hash
}

// signature R, S, V
type Signature struct {
	R *types.BigInt
	S *types.BigInt
	V *types.BigInt
}

func (s *Signature) init() {
	s.R = new(types.BigInt)
	s.S = new(types.BigInt)
	s.V = new(types.BigInt)
}

func (s *Signature)Init(){
	s.init()
}

type signValue interface {
	// check the signature obj is initialized, if not, throw painc
	checkObj()

	// get() computes R, S, V values corresponding to the
	// given signature.
	get(sig []byte) (err error)

	// convert to bytes
	toBytes() (sig []byte)
}

func (s *Signature) checkObj() {
	if s.R == nil || s.S == nil || s.V == nil {
		panic(fmt.Errorf("Signature obj is not initialized"))
	}
}

func (s *Signature)hashBytes()[]byte{
	srcBytes := s.toBytes()
	h := crypto.Keccak256(srcBytes)
	return h
}

func (s *Signature)Hash()types.Hash{
	srcBytes := s.toBytes()
	h := crypto.Keccak256(srcBytes)
	hash := types.Hash{}
	copy(hash[:] , h)
	return hash
}
func (s *Signature) get(sig []byte) (err error) {
	s.checkObj()

	if len(sig) != 65 {
		return errors.New(fmt.Sprintf("wrong size for Signature: got %d, want 65", len(sig)))
	} else {
		s.R.IntVal.SetBytes(sig[:32])
		s.S.IntVal.SetBytes(sig[32:64])

		if Config().chainId != nil && Config().chainId.Sign() != 0 {
			s.V.IntVal.SetInt64(int64(sig[64] + 35))
			s.V.IntVal.Add(&s.V.IntVal, Config().chainIdMul)
		} else {
			s.V.IntVal.SetBytes([]byte{sig[64] + 27})
		}
	}
	return nil
}

func (s Signature) toBytes() (sig []byte) {
	s.checkObj()

	sV := s.V
	V := types.BigInt{}
	if Config().chainId.Sign() != 0 {
		V.IntVal.Sub(&sV.IntVal, Config().chainIdMul)
		V.IntVal.Sub(&V.IntVal, common.Big35)
	} else{
		V.IntVal.Sub(&sV.IntVal, common.Big27)
	}

	vb := byte(V.IntVal.Uint64())
	if !crypto.ValidateSignatureValues(vb, &s.R.IntVal, &s.S.IntVal, true) {
		logger.Debugf("invalid Signature\n")
		return nil
	}

	rb, sb := s.R.IntVal.Bytes(), s.S.IntVal.Bytes()
	sig = make([]byte, 65)
	copy(sig[32-len(rb):32], rb)
	copy(sig[64-len(sb):64], sb)
	sig[64] = vb

	return sig
}

// long-term key singer
type CredentialSign struct {
	Signature
	Round  		uint64		// round
	Step   		uint64		// step
}

type CredentialSigForHash struct {
	Round  		uint64		// round
	Step   		uint64		// step
	Quantity    []byte		// quantity(seed, Qr-1)
}

func (a *CredentialSign) sigHashBig() *big.Int {
	h := a.Signature.hashBytes()
	return new(big.Int).SetBytes(h)
}

func (a *CredentialSign)Cmp(b *CredentialSign)int{
	h := a.Signature.hashBytes()
	aInt := new(big.Int).SetBytes(h)

	h = b.Signature.hashBytes()
	bInt := new(big.Int).SetBytes(h)

	return aInt.Cmp(bInt)
}

func (this *CredentialSign)ToCredentialSigKey()*CredentialSigForKey{
	r := new(CredentialSigForKey)
	r.Round = this.Round
	r.Step  = this.Step
	r.R     = this.R.IntVal.Uint64()
	r.S     = this.S.IntVal.Uint64()
	r.V     = this.V.IntVal.Uint64()
	return r
}
func (cret *CredentialSign)Sign(prv *ecdsa.PrivateKey)(R *types.BigInt, S *types.BigInt, V *types.BigInt, err error){
	return cret.sign(prv)
}
func (cret *CredentialSign) sign(prv *ecdsa.PrivateKey) (R *types.BigInt, S *types.BigInt, V *types.BigInt, err error) {
	if prv == nil {
		err := errors.New(fmt.Sprintf("private key is empty"))
		return nil, nil, nil, err
	}

	hash := cret.hash()
	if (hash == types.Hash{}) {
		err := errors.New(fmt.Sprintf("the hash of credential is empty"))
		return nil, nil, nil, err
	}

	sig, err := crypto.Sign(hash[:], prv)
	if err != nil {
		return nil, nil, nil, err
	}

	err = cret.get(sig)
	if err != nil {
		return nil, nil, nil, err
	}
	R = cret.R
	S = cret.S
	V = cret.V

	return R, S, V, nil
}

func (cret *CredentialSign) sender() (types.Address, error) {
	cret.checkObj()
	if Config().chainId != nil && deriveChainId(&cret.V.IntVal).Cmp(Config().chainId) != 0 {
		return types.Address{}, ErrInvalidChainId
	}
	if Config().chainId == nil{
		panic("Config().chainId == nil")
	}
	V := &big.Int{}
	if Config().chainId.Sign() != 0 {
		V = V.Sub(&cret.V.IntVal, Config().chainIdMul)
		V.Sub(V, common.Big35)
	} else{
		V = V.Sub(&cret.V.IntVal, common.Big27)
	}
	address, err :=  recoverPlain(cret.hash(), &cret.R.IntVal, &cret.S.IntVal, V, true)
	return address, err
}

func (cret *CredentialSign) hash() types.Hash {
	cretforhash := &CredentialSigForHash{
		cret.Round,
		cret.Step,
		[]byte{0},	// TODO: to get Quantity !!!!!!!!!!!!!!! need to implement a global function(round)
	}
	hash, err := common.MsgpHash(cretforhash)
	if err != nil {
		return types.Hash{}
	}
	return hash
}

// TODO: In current, EphemeralSig is the same as the Credential, need to be modified in the next version
// ephemeral key singer
type EphemeralSign struct {
	Signature
	round  		uint64		// round
	step   		uint64		// step
	val		    []byte		// Val = Hash(B), or Val = 0, or Val = 1
}

type EphemeralSigForHash struct {
	Round  		uint64		// round
	Step   		uint64		// step
	Val		    []byte		// Val = Hash(B), or Val = 0, or Val = 1
}

func (esig *EphemeralSign)GetStep()uint64{
	return esig.step
}
func (esig *EphemeralSign) Sign(prv *ecdsa.PrivateKey) (R *types.BigInt, S *types.BigInt, V *types.BigInt, err error) {
	return esig.sign(prv)
}

func (esig *EphemeralSign) sign(prv *ecdsa.PrivateKey) (R *types.BigInt, S *types.BigInt, V *types.BigInt, err error) {
	if prv == nil {
		err := errors.New(fmt.Sprintf("private key is empty"))
		return nil, nil, nil, err
	}

	hash := esig.hash()
	if (hash == types.Hash{}) {
		err := errors.New(fmt.Sprintf("the hash of credential is empty"))
		return nil, nil, nil, err
	}

	sig, err := crypto.Sign(hash[:], prv)
	if err != nil {
		return nil, nil, nil, err
	}

	err = esig.get(sig)
	if err != nil {
		return nil, nil, nil, err
	}
	R = esig.R
	S = esig.S
	V = esig.V

	return R, S, V, nil
}

func (esig *EphemeralSign) sender() (types.Address, error) {
	esig.checkObj()

	if Config().chainId != nil && deriveChainId(&esig.V.IntVal).Cmp(Config().chainId) != 0 {
		return types.Address{}, ErrInvalidChainId
	}

	V := &big.Int{}
	if Config().chainId.Sign() != 0 {
		V = V.Sub(&esig.V.IntVal, Config().chainIdMul)
		V.Sub(V, common.Big35)
	} else{
		V = V.Sub(&esig.V.IntVal, common.Big27)
	}
	address, err :=  recoverPlain(esig.hash(), &esig.R.IntVal, &esig.S.IntVal, V, true)
	return address, err
}

func (esig *EphemeralSign) hash() types.Hash {
	if esig.val == nil {
		panic(fmt.Errorf("EphemeralSign obj is not initialized"))
	}
	eisgforhash := &EphemeralSigForHash{
		esig.round,
		esig.step,
		esig.val,
	}
	hash, err := common.MsgpHash(eisgforhash)
	if err != nil {
		return types.Hash{}
	}
	return hash
}

func RecoverPlain(sighash types.Hash, R, S, Vb *big.Int, homestead bool) (types.Address, error) {
	return recoverPlain(sighash , R,S,Vb , homestead)
}
func recoverPlain(sighash types.Hash, R, S, Vb *big.Int, homestead bool) (types.Address, error) {
	if Vb.BitLen() > 8 {
		return types.Address{}, ErrInvalidSig
	}
	V := byte(Vb.Uint64())
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return types.Address{}, ErrInvalidSig
	}
	// encode the snature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the signature
	pub, err := crypto.Ecrecover(sighash[:], sig)
	if err != nil {
		return types.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return types.Address{}, errors.New("invalid public key")
	}
	var addr types.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return addr, nil
}

// deriveChainId derives the chain id from the given v parameter
func deriveChainId(v *big.Int) *big.Int {
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}
