package internal

import (
	"fmt"
	"testing"
	"unicode/utf8"
)

func Test_MatchingKetIndex(t *testing.T) {
	var tt = []struct {
		name        string
		inputString string
		braIndex    int
		expected    int
	}{
		{"", "()()", 0, 1},
		{"", "()()", 2, 3},
		{"", "()()", 1, -1},
		{"", "(((),()))", 0, 8},
		{"", "(((),()))", 1, 7},
		{"", "(((),()))", 2, 3},
		{"", "aasdfmul(123,mul(543,24))", 8, 24},

		{"", "[\u001B]", 0, 2},
		{"", "[\"]", 0, 2},

		//These are a failing tests. Left here for informative purposes only. They pass if tested with
		//method matchingKetIndexRune.
		{"", "[�]", 0, 2},
		{"", "[·]", 0, 2},
		{"", "[map[extra_attributes:map[average_color:Y�c�Pe�@\u001B�^)#@�@m{�Z.��@ img:/9j/2wCEAAgGBgcGBQgHBw] id:img:0.0.0.0:1 values:[]]]", 0, 100},
		{"", "[map[�Pe�@\u001B�]]", 0, 13},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ketIndex := matchingKetIndex([]byte(tc.inputString), tc.braIndex)
			if ketIndex != tc.expected {
				t.Errorf("ketIndex => %d, want %d", ketIndex, tc.expected)
			}
		})
	}
}

func TestBracketStack(t *testing.T) {
	var tt = []struct {
		name        string
		inputString string
	}{
		{"", "()()"},
		{"", "(((),()))"},
		{"", "aasdfmul(123,mul(543,24))"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}

func TestBracketsPush(t *testing.T) {
	var tt = []struct {
		name                string
		inputBytes          []byte
		expectedBraStackLen int
		expectedBraIndexLen int
		expectedValid       int
	}{
		{"", []byte("((((((("), 7, 7, 0},
		{"", []byte("()()"), 0, 0, 2},
		{"", []byte("))()()(("), 2, 2, 2},
		{"", []byte("())()()(()"), 1, 1, 4},
		{"", []byte("(()"), 1, 1, 1},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			//cs := newBracketMatcher()
			//for i, v := range tc.inputBytes {
			//	cs.push(v, i)
			//}
			//
			//if len(cs.matchedBrackets) != tc.expectedValid {
			//	t.Errorf("got %v, expected %v valid bracketMatcher", len(cs.matchedBrackets), tc.expectedValid)
			//}
			//if len(cs.braStack) != tc.expectedBraStackLen {
			//	t.Errorf("got %v, expected %v bra stack length", len(cs.braStack), tc.expectedBraStackLen)
			//}
			//if len(cs.braIndexStack) != tc.expectedBraIndexLen {
			//	t.Errorf("got %v, expected %v bra index stack length", len(cs.braIndexStack), tc.expectedBraIndexLen)
			//}
		})
	}
}

func TestBraMatches(t *testing.T) {
	var tt = []struct {
		name     string
		input    byte
		match    byte
		expected bool
	}{
		{"", '(', ')', true},
		{"", '(', '}', false},
		{"", '(', '(', false},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			br := bra(tc.input)
			ans := br.matches(tc.match)
			if ans != tc.expected {
				t.Errorf("got %v, expected %v", ans, tc.expected)
			}
		})
	}
}

func Test_isBra(t *testing.T) {
	var tt = []struct {
		name     string
		input    byte
		expected bool
	}{
		{"", '(', true},
		{"", '[', true},
		{"", ']', false},
		{"", 'a', false},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ans := isBra(tc.input)
			if ans != tc.expected {
				t.Errorf("got %v, expected %v", ans, tc.expected)
			}
		})
	}
}

func Test_isBraRune(t *testing.T) {
	var tt = []struct {
		name     string
		input    rune
		expected bool
	}{
		{"", '(', true},
		{"", '[', true},
		{"", ']', false},
		{"", 'a', false},
		{"", '‡', false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ans := isBraRune(tc.input)
			if ans != tc.expected {
				t.Errorf("got %v, expected %v", ans, tc.expected)
			}
		})
	}
}

func Test_NotATest(t *testing.T) {
	s := "[A☕あ]" // ASCII, emoji, Japanese character

	fmt.Println("String:", s)
	fmt.Println("Bytes:", len(s))
	fmt.Println("Runes:", utf8.RuneCountInString(s))
	fmt.Println()

	fmt.Println("Byte by byte:")
	for i := 0; i < len(s); i++ {
		fmt.Printf("  byte[%d] = %x\n", i, s[i])
	}
	fmt.Println()

	fmt.Println("Rune by rune:")
	for i, r := range s {
		fmt.Printf("  rune[%d] = %c (U+%04X)\n", i, r, r)
	}
}

func Test_NotATest2(t *testing.T) {
	s := "A☕あ"

	fmt.Println("String:", s)
	fmt.Printf("Total bytes: %d\n", len(s))
	fmt.Printf("Total runes: %d\n\n", utf8.RuneCountInString(s))

	// Indexing by byte
	fmt.Println("Indexing by byte:")
	for i := 0; i < len(s); i++ {
		fmt.Printf("  s[%d] = %x (%q)\n", i, s[i], s[i])
	}
	fmt.Println()

	// Iterating by rune
	fmt.Println("Iterating by rune (using for range):")
	for i, r := range s {
		fmt.Printf("  at byte %d: rune = %c (U+%04X)\n", i, r, r)
	}
}

func Test_matchingKetIndexRune(t *testing.T) {
	var tt = []struct {
		name        string
		inputString string
		braIndex    int
		expected    int
	}{
		{"", "(((),()))", 2, 3},
		{"", "aasdfmul(123,mul(543,24))", 8, 24},
		{"", "[map[extra_attributes:map[average_color:Y�c�Pe�@\u001B�^)#@�@m{�Z.��@ img:/9j/2wCEAAgGBgcGBQgHBw] id:img:0.0.0.0:1 values:[]]]", 0, 120},
		{"", "[\u001B]", 0, 2},
		{"", "[\u001B\u001B\u001B]", 0, 4},
		{"", "[\"]", 0, 2},
		{"", "[�]", 0, 2},
		{"", "[·]", 0, 2},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s := []rune(tc.inputString)
			ketIndex := matchingKetIndexRune(s, tc.braIndex)
			if ketIndex != tc.expected {
				t.Errorf("ketIndex => %d, want %d", ketIndex, tc.expected)
			}
		})
	}
}
