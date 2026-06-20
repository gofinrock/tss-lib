// Copyright © 2019 Binance
//
// This file is part of Binance. The full Binance copyright notice, including
// terms governing use, modification, and redistribution, is contained in the
// file LICENSE at the root of the source code distribution tree.

package resharing

import (
	"errors"

	"github.com/bnb-chain/tss-lib/v4/tss"
)

func (round *round2) Start() *tss.Error {
	if round.started {
		return round.WrapError(errors.New("round already started"))
	}
	round.number = 2
	round.started = true
	round.resetOK() // resets both round.oldOK and round.newOK
	round.allOldOK()

	if !round.ReSharingParams().IsNewCommittee() {
		return nil
	}
	// NOTE(patch): round2.Update() gates the OLD committee on newOK. Pre-marking
	// allNewOK() for a party in BOTH committees (refresh) defeats that gate, so it
	// advances before collecting the new committee's DGRound2 messages and
	// deadlocks. New-only parties are not old-committee receivers, so they may
	// still pre-mark. (ecdsa/resharing round 2 has no allNewOK() call at all.)
	if !round.ReSharingParams().IsOldCommittee() {
		round.allNewOK()
	}

	Pi := round.PartyID()
	i := Pi.Index

	// 1. "broadcast" "ACK" members of the OLD committee
	r2msg := NewDGRound2Message(round.OldParties().IDs(), Pi)
	round.temp.dgRound2Messages[i] = r2msg
	round.out <- r2msg

	return nil
}

func (round *round2) CanAccept(msg tss.ParsedMessage) bool {
	if _, ok := msg.Content().(*DGRound2Message); ok {
		return msg.IsBroadcast()
	}
	return false
}

func (round *round2) Update() (bool, *tss.Error) {
	// only the old committee receive in this round
	if !round.ReSharingParams().IsOldCommittee() {
		return true, nil
	}

	ret := true
	// accept messages from new -> old committee
	for j, msg := range round.temp.dgRound2Messages {
		if round.newOK[j] {
			continue
		}
		if msg == nil || !round.CanAccept(msg) {
			ret = false
			continue
		}
		round.newOK[j] = true
	}

	return ret, nil
}

func (round *round2) NextRound() tss.Round {
	round.started = false
	return &round3{round}
}
