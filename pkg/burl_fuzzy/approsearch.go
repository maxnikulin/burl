// Copyright (C) 2021 Max Nikulin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

/*
Approximate sub-string matching using edit distance.
Damerau-Levenshtein distance (with transposions/swaps)
*/
package burl_fuzzy

import (
	// "fmt" // debug output
	"strings"
)

type EditorDistance int

const (
	TranspositionDistance = EditorDistance(1)
	DeletionDistance      = EditorDistance(1)
	InsertionDistance     = EditorDistance(1)
	SubstitutionDistance  = EditorDistance(1)
	CutCost               = EditorDistance(32767)
)

func SameRuneDistance(p rune, w rune) EditorDistance {
	if p == w {
		return EditorDistance(0)
	} else {
		return EditorDistance(1)
	}
}

func MinDistance(args ...EditorDistance) (result EditorDistance) {
	result = CutCost
	for _, v := range args {
		if v < result {
			result = v
		}
	}
	return
}

func EditorDistanceRunes(pattern []rune, word []rune) EditorDistance {
	// fmt.Printf("%2c -----\n", word);

	// query may start at an arbitrary position in word so
	// we are starting with zero distance, not from
	//   (i + 1)*InsertionCost
	// as in the classical match word approach in Wagner-Fischer algorithm.
	costAlt := make([]EditorDistance, len(word))
	for i := 0; i < len(word); i++ {
		costAlt[i] = 0
	}

	cost := make([]EditorDistance, len(word))
	// Minor optimization assuming costAlt[i] = 0
	initDistance := MinDistance(SubstitutionDistance, DeletionDistance)
	for i := 0; i < len(word); i++ {
		cost[i] = initDistance * SameRuneDistance(pattern[0], word[i])
	}
	// fmt.Printf("%2v %c\n", cost, pattern[0])

	// Transpositions require d[i - 2, j - 2] value, actually it is enough
	// to remember just 2 values from i - 2 row.
	var prevDistance [2]EditorDistance

	for iPattern := 1; iPattern < len(pattern); iPattern++ {
		cost, costAlt = costAlt, cost
		prevDistance[1] = DeletionDistance * EditorDistance(iPattern-1)
		prevDistance[0] = cost[0]
		cost[0] = MinDistance(
			SubstitutionDistance*SameRuneDistance(pattern[iPattern], word[0])+
				DeletionDistance*EditorDistance(iPattern),
			InsertionDistance+DeletionDistance*EditorDistance(iPattern),
			costAlt[0]+DeletionDistance,
		)
		for iWord := 1; iWord < len(word); iWord++ {
			d_transposition := CutCost
			if pattern[iPattern-1] == word[iWord] && pattern[iPattern] == word[iWord-1] {
				d_transposition = prevDistance[iWord%2] + TranspositionDistance
			}
			prevDistance[iWord%2] = cost[iWord]

			cost[iWord] = MinDistance(
				costAlt[iWord-1]+SubstitutionDistance*SameRuneDistance(pattern[iPattern], word[iWord]),
				cost[iWord-1]+InsertionDistance,
				costAlt[iWord]+DeletionDistance,
				d_transposition,
			)
		}
		// fmt.Printf("%2v %c\n", cost, pattern[iPattern])
	}
	return MinDistance(cost...)
}

func SearchDistanceDL(pattern string, word string) EditorDistance {
	patternRunes := []rune(strings.ToLower(pattern))
	wordRunes := []rune(strings.ToLower(word))
	return EditorDistanceRunes(patternRunes, wordRunes)
}
