// Package printer holds a printer utility class for printing indented "pretty print" style output.
package printer

import (
	"fmt"
	"io"
)

const (
	indentAmount = 4
)

// Printer is a helper struct meant to pretty print the frontend data.
type Printer struct {
	indent int
	w      io.Writer
}

func NewPrinter(w io.Writer) *Printer {
	return &Printer{
		w: w,
	}
}

func (p *Printer) WriteLine(format string, args ...interface{}) {
	result := fmt.Sprintf(format, args...)
	final := fmt.Sprintf("%*s%s\n", p.indent, " ", result)
	p.w.Write([]byte(final))
}

func (p *Printer) PushIndent() {
	p.indent += indentAmount
}

func (p *Printer) PopIndent() {
	p.indent -= indentAmount
	if p.indent < 0 {
		panic("Unmatched popIndent")
	}
}
