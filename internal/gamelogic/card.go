package gamelogic

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
	SuitYellow
	SuitOrange
)

var Suits = []Suit{SuitRed, SuitGreen, SuitBlue, SuitPurple, SuitYellow, SuitOrange}

var suitName = map[Suit]string{
	SuitRed:    "red",
	SuitGreen:  "green",
	SuitBlue:   "blue",
	SuitPurple: "purple",
	SuitYellow: "yellow",
	SuitOrange: "orange",
}

func (s Suit) String() string {
	return suitName[s]
}

type Card struct {
	Suit  Suit `json:"suit"`
	Value int  `json:"value"`
}

func (c Card) String() string {
	return fmt.Sprintf("%s %d", strings.ToUpper(c.Suit.String())[:1], c.Value)
}

func (c Card) SuitSortingValue() int {
	return int(c.Suit)*100 + (c.Value - 1)
}
