// Copyright © 2019 Binance
//
// This file is part of Binance. The full Binance copyright notice, including
// terms governing use, modification, and redistribution, is contained in the
// file LICENSE at the root of the source code distribution tree.

package common_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/tss-lib/v4/common"
)

const (
	randomIntBitLen = 1024
)

func TestGetRandomInt(t *testing.T) {
	rnd := common.MustGetRandomInt(rand.Reader, randomIntBitLen)
	assert.NotZero(t, rnd, "rand int should not be zero")
}

func TestGetRandomPositiveInt(t *testing.T) {
	rnd := common.MustGetRandomInt(rand.Reader, randomIntBitLen)
	rndPos := common.GetRandomPositiveInt(rand.Reader, rnd)
	assert.NotZero(t, rndPos, "rand int should not be zero")
	assert.True(t, rndPos.Cmp(big.NewInt(0)) == 1, "rand int should be positive")
}

func TestGetRandomPositiveIntBoundaries(t *testing.T) {
	t.Run("nil lessThan returns nil", func(tt *testing.T) {
		assert.Nil(tt, common.GetRandomPositiveInt(rand.Reader, nil))
	})
	t.Run("lessThan == 0 returns nil", func(tt *testing.T) {
		assert.Nil(tt, common.GetRandomPositiveInt(rand.Reader, big.NewInt(0)))
	})
	t.Run("lessThan == 1 returns nil (no value in (0, 1))", func(tt *testing.T) {
		// Old implementation would loop forever sampling 0; new implementation
		// rejects up-front because (0, 1) is empty.
		assert.Nil(tt, common.GetRandomPositiveInt(rand.Reader, big.NewInt(1)))
	})
	t.Run("lessThan == 2 always returns 1", func(tt *testing.T) {
		// (0, 2) = {1} — sampler must return 1, never 0.
		for i := 0; i < 20; i++ {
			v := common.GetRandomPositiveInt(rand.Reader, big.NewInt(2))
			assert.NotNil(tt, v)
			assert.Equal(tt, int64(1), v.Int64())
		}
	})
}

func TestGetRandomPositiveRelativelyPrimeInt(t *testing.T) {
	rnd := common.MustGetRandomInt(rand.Reader, randomIntBitLen)
	rndPosRP := common.GetRandomPositiveRelativelyPrimeInt(rand.Reader, rnd)
	assert.NotZero(t, rndPosRP, "rand int should not be zero")
	assert.True(t, common.IsNumberInMultiplicativeGroup(rnd, rndPosRP))
	assert.True(t, rndPosRP.Cmp(big.NewInt(0)) == 1, "rand int should be positive")
	// TODO test for relative primeness
}

func TestGetRandomPrimeInt(t *testing.T) {
	prime := common.GetRandomPrimeInt(rand.Reader, randomIntBitLen)
	assert.NotZero(t, prime, "rand prime should not be zero")
	assert.True(t, prime.ProbablyPrime(50), "rand prime should be prime")
}
