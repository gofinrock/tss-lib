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

func (round *round3) Start() *tss.Error {
	if round.started {
		return round.WrapError(errors.New("round already started"))
	}
	round.number = 3
	round.started = true
	round.resetOK() // resets both round.oldOK and round.newOK
	round.allNewOK()

	if !round.ReSharingParams().IsOldCommittee() {
		return nil
	}
	// NOTE(patch): see round_1_old_step_1.go. round3.Update() gates the new
	// committee on oldOK; pre-marking allOldOK() for a party in both committees
	// (refresh) lets it advance to round 4 before collecting the old committee's
	// DGRound3 messages, aborting with "dgRound3Message2 not received". Old-only
	// parties do not receive in this round, so they must still pre-mark.
	if !round.ReSharingParams().IsNewCommittee() {
		round.allOldOK()
	}

	Pi := round.PartyID()
	i := Pi.Index

	// 2. send share to Pj from the new committee
	for j, Pj := range round.NewParties().IDs() {
		share := round.temp.NewShares[j]
		r3msg1 := NewDGRound3Message1(Pj, round.PartyID(), share)
		// NOTE(patch): store ONLY our own share in the self slot, and do not
		// emit it to the wire. The original code stored r3msg1 into
		// dgRound3Message1s[i] on every iteration, so the self slot ended up
		// holding the share addressed to the LAST new party instead of our own.
		// Upstream masks this because the self share is also sent over the
		// network and received back (overwriting the bad value); a driver that
		// skips self-delivery is left with a wrong self share that fails
		// Verify(). Detect self by identity (KeyInt), not index, since old/new
		// committee indices can collide for overlapping committees.
		if Pj.KeyInt().Cmp(round.PartyID().KeyInt()) == 0 {
			round.temp.dgRound3Message1s[i] = r3msg1
			continue
		}
		round.out <- r3msg1
	}

	vDeCmt := round.temp.VD
	r3msg2 := NewDGRound3Message2(
		round.NewParties().IDs().Exclude(round.PartyID()), round.PartyID(),
		vDeCmt)
	round.temp.dgRound3Message2s[i] = r3msg2
	round.out <- r3msg2

	return nil
}

func (round *round3) CanAccept(msg tss.ParsedMessage) bool {
	if _, ok := msg.Content().(*DGRound3Message1); ok {
		return !msg.IsBroadcast()
	}
	if _, ok := msg.Content().(*DGRound3Message2); ok {
		return msg.IsBroadcast()
	}
	return false
}

func (round *round3) Update() (bool, *tss.Error) {
	// only the new committee receive in this round
	if !round.ReSharingParams().IsNewCommittee() {
		return true, nil
	}
	// accept messages from old -> new committee
	for j, msg1 := range round.temp.dgRound3Message1s {
		if round.oldOK[j] {
			continue
		}
		if msg1 == nil || !round.CanAccept(msg1) {
			return false, nil
		}
		msg2 := round.temp.dgRound3Message2s[j]
		if msg2 == nil || !round.CanAccept(msg2) {
			return false, nil
		}
		round.oldOK[j] = true
	}
	return true, nil
}

func (round *round3) NextRound() tss.Round {
	round.started = false
	return &round4{round}
}
