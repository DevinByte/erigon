// Copyright 2021 The Erigon Authors
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package cmp

import (
	"golang.org/x/exp/constraints"
)

// InRange - ensure val is in [min,max] range
func InRange[T constraints.Ordered](min, max, val T) T {
	if min >= val {
		return min
	}
	if max <= val {
		return max
	}
	return val
}

func Compare[T constraints.Ordered](a, b T) int {
	switch {
	case a < b:
		return -1
	case a == b:
		return 0
	default:
		return 1
	}
}
