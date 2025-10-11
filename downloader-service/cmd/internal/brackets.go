package internal

import (
	"unicode/utf8"
)

type bra byte
type ket byte

const (
	Bra40 bra = '('
	Ket41 ket = ')'
	Bra91 bra = '['
	Ket93 ket = ']'
)

//func FindMatchingBrackets(input []byte) map[int]int {
//	matcher := bracketStack{}
//	for i, b := range input {
//		matcher.push(b)
//	}
//	return matcher.matchedBrackets
//}

var matchingBrackets = map[bra]ket{
	Bra40: Ket41,
	Bra91: Ket93,
}

func (b bra) matches(k byte) bool {
	if expected, ok := matchingBrackets[b]; ok {
		return byte(expected) == k
	}
	return false
}

func isBra(b byte) bool {
	_, ok := matchingBrackets[bra(b)]
	return ok
}

func isBraRune(r rune) bool {
	b := byte(r)
	_, ok := matchingBrackets[bra(b)]
	return ok
}

func isKet(b byte) bool {
	for _, k := range matchingBrackets {
		if byte(k) == b {
			return true
		}
	}
	return false
}

type bracketStack []bra

func (bs *bracketStack) push(b byte) int {
	if isBra(b) {
		*bs = append(*bs, bra(b))
	} else if len(*bs) > 0 && isKet(b) {
		lastBra := (*bs)[len(*bs)-1]
		if lastBra.matches(b) {
			*bs = (*bs)[:len(*bs)-1]
		}
	}
	return len(*bs)
}

/* TIP Fails if the byte slice originates from a string of characters >= utf8.RuneSelf*/
func matchingKetIndex(s []byte, index int) int {
	if !isBra(s[index]) {
		return -1
	}

	stack := bracketStack(make([]bra, 0))

	matchingKetIndx := index
	depth := 0
	for i, c := range s[index:] {
		depth = stack.push(c)
		if depth == 0 {
			matchingKetIndx += i
			break
		}
	}

	if matchingKetIndx == index {
		return -1
	}

	return matchingKetIndx
}

func matchingKetIndexRune(s []rune, index int) int {
	r := s[index]
	if r >= utf8.RuneSelf {
		return -1
	}
	b := byte(r)
	if !isBra(b) {
		return -1
	}

	stack := bracketStack(make([]bra, 0))
	matchingKetIndx := index
	depth := 0
	for i, c := range s[index:] {
		if c >= utf8.RuneSelf {
			continue
		}
		depth = stack.push(byte(c))
		if depth == 0 {
			matchingKetIndx += i
			break
		}
	}

	if matchingKetIndx == index {
		return -1
	}

	return matchingKetIndx
}
