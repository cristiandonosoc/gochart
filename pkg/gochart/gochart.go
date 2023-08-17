// Package gochart is the main entrypoint to the compiler as a library.
// For using it, see cmd/gochart
package gochart

import (
	"fmt"
	"io"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/backend"
)

type Configuration struct {
	Frontend frontend.FrontendKind
	FrontendOptions []frontend.FrontendOption

	Backend backend.BackendKind
	BackendOptions []backend.BackendOption
}

type gochartCompiler struct {
	conf *Configuration
}

func NewGochartCompiler(conf *Configuration) (*gochartCompiler, error)  {
	return nil, fmt.Errorf("IMPLEMENT ME")
}

func (gc *gochartCompiler) PrintFrontend() (io.Reader, error) {
	return nil, fmt.Errorf("IMPLEMENT ME")

}

func (gc *gochartCompiler) PrintIR(w io.Writer) (io.Reader, error) {
	return nil, fmt.Errorf("IMPLEMENT ME")
}

type BackendOutput struct {
	Name string
	Output io.Reader
}

func (gc *gochartCompiler) PrintBackend(w io.Writer) ([]*BackendOutput, error) {
	return nil, fmt.Errorf("IMPLEMENT ME")
}




