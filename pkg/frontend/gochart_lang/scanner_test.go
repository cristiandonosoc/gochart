package gochart_lang

import (
	"testing"
	"strings"

	"github.com/stretchr/testify/assert"
)

func TestGatherTokens(t *testing.T) {
	input := `
// This is a comment.
(( )){} // Grouping stuff.
[[]]
	`

	s := NewScanner()

	r := strings.NewReader(input)
	_, errors := s.Scan(r)
	assert.Empty(t, errors)
}
