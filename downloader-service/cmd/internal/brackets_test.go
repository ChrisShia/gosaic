package internal

import "testing"

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

func TestValidBra(t *testing.T) {
	var tt = []struct {
		name     string
		input    byte
		expected bool
	}{
		{"", '(', true},
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
