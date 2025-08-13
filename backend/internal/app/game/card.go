package game

import (
	"fmt"
	"strings"
)

type Suit int

const (
	SuitRed Suit = iota
	SuitGreen
	SuitBlue
	SuitPurple
	SuitBlack
)

var Suits = []Suit{SuitRed, SuitGreen, SuitBlue, SuitPurple, SuitBlack}

var suitName = map[Suit]string{
	SuitRed:    "red",
	SuitGreen:  "green",
	SuitBlue:   "blue",
	SuitPurple: "purple",
	SuitBlack:  "black",
}

func (s Suit) String() string {
	return suitName[s]
}

type Card struct {
	Suit  Suit `json:"suit"`
	Value int  `json:"value"`
}

func (c Card) String() string {
	return fmt.Sprintf("%s %d", strings.ToUpper(c.Suit.String()), c.Value)
}

func (c Card) SuitSortingValue() int {
	return int(c.Suit)*100 + (c.Value - 1)
}
