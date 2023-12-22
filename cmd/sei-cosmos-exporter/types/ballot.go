package types

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	cosmos_math "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NOTE: we don't need to implement proto interface on this file
//       these are not used in store or rpc response

// VoteForTally is a convenience wrapper to reduce redundant lookup cost
type VoteForTally struct {
	Denom        string
	ExchangeRate cosmos_math.LegacyDec
	Voter        sdk.ValAddress
	Power        int64
}

// NewVoteForTally returns a new VoteForTally instance
func NewVoteForTally(rate cosmos_math.LegacyDec, denom string, voter sdk.ValAddress, power int64) VoteForTally {
	return VoteForTally{
		ExchangeRate: rate,
		Denom:        denom,
		Voter:        voter,
		Power:        power,
	}
}

// ExchangeRateBallot is a convenience wrapper around a ExchangeRateVote slice
type ExchangeRateBallot []VoteForTally

// ToMap return organized exchange rate map by validator
func (pb ExchangeRateBallot) ToMap() map[string]cosmos_math.LegacyDec {
	exchangeRateMap := make(map[string]cosmos_math.LegacyDec)
	for _, vote := range pb {
		if vote.ExchangeRate.IsPositive() {
			exchangeRateMap[string(vote.Voter)] = vote.ExchangeRate
		}
	}

	return exchangeRateMap
}

// ToCrossRate return cross_rate(base/exchange_rate) ballot
func (pb ExchangeRateBallot) ToCrossRate(bases map[string]cosmos_math.LegacyDec) (cb ExchangeRateBallot) {
	for i := range pb {
		vote := pb[i]

		if exchangeRateRT, ok := bases[string(vote.Voter)]; ok && vote.ExchangeRate.IsPositive() {
			vote.ExchangeRate = exchangeRateRT.Quo(vote.ExchangeRate)
		} else {
			// If we can't get reference Sei exchange rate, we just convert the vote as abstain vote
			vote.ExchangeRate = cosmos_math.LegacyZeroDec()
			vote.Power = 0
		}

		cb = append(cb, vote)
	}

	return
}

// ToCrossRateWithSort return cross_rate(base/exchange_rate) ballot
func (pb ExchangeRateBallot) ToCrossRateWithSort(bases map[string]cosmos_math.LegacyDec) (cb ExchangeRateBallot) {
	cb = pb.ToCrossRate(bases)
	sort.Sort(cb)
	return
}

// Power returns the total amount of voting power in the ballot
func (pb ExchangeRateBallot) Power() int64 {
	totalPower := int64(0)
	for _, vote := range pb {
		totalPower += vote.Power
	}

	return totalPower
}

// WeightedMedian returns the median weighted by the power of the ExchangeRateVote.
// CONTRACT: ballot must be sorted
func (pb ExchangeRateBallot) WeightedMedian() cosmos_math.LegacyDec {
	totalPower := pb.Power()
	if pb.Len() > 0 {
		pivot := int64(0)
		for _, v := range pb {
			votePower := v.Power

			pivot += votePower
			if pivot >= (totalPower / 2) {
				return v.ExchangeRate
			}
		}
	}
	return cosmos_math.LegacyZeroDec()
}

// WeightedMedianWithAssertion returns the median weighted by the power of the ExchangeRateVote.
// CONTRACT: ballot must be sorted
func (pb ExchangeRateBallot) WeightedMedianWithAssertion() cosmos_math.LegacyDec {
	if !sort.IsSorted(pb) {
		panic("ballot must be sorted")
	}

	totalPower := pb.Power()
	if pb.Len() > 0 {
		pivot := int64(0)
		for _, v := range pb {
			votePower := v.Power

			pivot += votePower
			if pivot >= (totalPower / 2) {
				return v.ExchangeRate
			}
		}
	}
	return cosmos_math.LegacyZeroDec()
}

// StandardDeviation returns the standard deviation by the power of the ExchangeRateVote.
func (pb ExchangeRateBallot) StandardDeviation(median cosmos_math.LegacyDec) (standardDeviation cosmos_math.LegacyDec) {
	if len(pb) == 0 {
		return cosmos_math.LegacyZeroDec()
	}

	defer func() {
		if e := recover(); e != nil {
			standardDeviation = cosmos_math.LegacyZeroDec()
		}
	}()

	sum := cosmos_math.LegacyZeroDec()
	for _, v := range pb {
		deviation := v.ExchangeRate.Sub(median)
		sum = sum.Add(deviation.Mul(deviation))
	}

	variance := sum.QuoInt64(int64(len(pb)))

	floatNum, _ := strconv.ParseFloat(variance.String(), 64)
	floatNum = math.Sqrt(floatNum)
	standardDeviation, _ = cosmos_math.LegacyNewDecFromStr(fmt.Sprintf("%f", floatNum))

	return
}

// Len implements sort.Interface
func (pb ExchangeRateBallot) Len() int {
	return len(pb)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (pb ExchangeRateBallot) Less(i, j int) bool {
	return pb[i].ExchangeRate.LT(pb[j].ExchangeRate)
}

// Swap implements sort.Interface.
func (pb ExchangeRateBallot) Swap(i, j int) {
	pb[i], pb[j] = pb[j], pb[i]
}

// Claim is an interface that directs its rewards to an attached bank account.
type Claim struct {
	Power     int64
	Weight    int64
	WinCount  int64
	DidVote   bool
	Recipient sdk.ValAddress
}

// NewClaim generates a Claim instance.
func NewClaim(power, weight, winCount int64, recipient sdk.ValAddress, didVote bool) Claim {
	return Claim{
		Power:     power,
		Weight:    weight,
		WinCount:  winCount,
		DidVote:   didVote,
		Recipient: recipient,
	}
}
