package picsum

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func Random200300() *http.Request {
	return &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   "picsum.photos",
			Path:   "/200/300",
		},
	}
}

type PicUrlParameters struct {
	X          int
	Y          int
	Seed       string
	Signifiers []string
	Attributes Attributes
	Format     string
}

func (p *PicUrlParameters) Path() (s string) {
	builder := strings.Builder{}
	builder.WriteString("/")
	builder.WriteString(strconv.Itoa(p.X))
	if p.Y != p.X {
		builder.WriteString("/")
		builder.WriteString(strconv.Itoa(p.Y))
	}
	builder.WriteString("/")
	builder.WriteString(p.Attributes.String())
	builder.WriteString(p.Format)
	return builder.String()
}

func NewAttributes(m map[string]string) Attributes {
	as := make([]Attribute, 0)
	for key, value := range m {
		attr := Attribute{Name: key, Value: value}
		as = append(as, attr)
	}
	return as
}

type Attribute struct {
	Name  string
	Value string
}

func (a *Attribute) String() string {
	return fmt.Sprintf("%s=%s", a.Name, a.Value)
}

type Attributes []Attribute

func (a Attributes) StringPlain() string {
	return fmt.Sprintf("%#v", a)
}

func (a Attributes) String() string {
	stringBuilder := strings.Builder{}
	for _, attr := range a {
		stringBuilder.WriteString(attr.String())
		stringBuilder.WriteString("&")
	}
	res := stringBuilder.String()
	res = strings.TrimRight(res, "&")
	return res
}
