// Package frontend holds the interface for reading different kind of languages/specs that we might
// support for reading statecharts.
package frontend

import (
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
)

// GochartFrontend is the abstract interface for all frontends, regardless of the type of data they
// consume. This is used to decouple language specifications from the rest of the program.
type GochartFrontend interface {

}

// NewGochartLangFrontend returns a frontend capable of parting the gochart language.
func NewGochartLangFrontend() GochartFrontend {
	return gochart_lang.NewFrontend()
}
