package picsum

import (
	"testing"
)

func Test_AttributeString(t *testing.T) {
	var tt = []struct {
		attribute Attribute
		expected  string
	}{
		{Attribute{Name: "blur", Value: "2"}, "blur=2"},
	}

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			actual := tc.attribute.String()
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}

func Test_AttributesString(t *testing.T) {
	var tt = []struct {
		attributes Attributes
		expected   string
	}{
		{a("A1", "V1", "A2", "V2", "A3", "V3"), "A1=V1&A2=V2&A3=V3"},
	}

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			actual := tc.attributes.String()
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}

func a(s ...string) Attributes {
	m := make(map[string]string)
	for i := 0; i < len(s); i = i + 2 {
		m[s[i]] = s[i+1]
	}
	return NewAttributes(m)
}
