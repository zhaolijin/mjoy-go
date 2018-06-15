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
// @File: types.go
// @Date: 2018/06/12 11:01:51
////////////////////////////////////////////////////////////////////////////////

package apos

import (
	"sync"
	"bytes"
	"github.com/tinylib/msgp/msgp"
	"math/big"
)

//some system param(algorand system param) for step goroutine.
type Config struct {
	Lookback            uint32		// lookback val, r - k
	PrPrecision			*big.Int	// the precision
	PrLeader			*big.Int	// the probability of Leaders
	PrVerifier			*big.Int	// the probability of Verifiers
	MaxBBASteps         uint32		// the max number of BBA steps
	MaxNumPerRound      uint32		// the max number of nodes per round
	PrH					*big.Int	// the probability of honest
	DelayBlock          int  		// time A, sec
	DelayVerify         int  		// time λ, sec
}

func (c *Config) setDefault() {
	c.Lookback = 100
	c.PrPrecision = big.NewInt(10)
	c.PrLeader = big.NewInt(1000000000)		// 0.1
	c.PrVerifier = big.NewInt(5000000000) 	// 0.5
	c.MaxBBASteps = 180
	c.MaxNumPerRound = 10
	c.PrH = big.NewInt(34)
	c.DelayBlock = 60
	c.DelayVerify = 10
}
