package netscapefmt

import "strings"

type ncPrinter struct {
	value string
	depth int
}

func (p *ncPrinter) appendLine(v string) {
	p.value += strings.Repeat(" ", p.depth) + v + "\n"
}

func (p *ncPrinter) inc() {
	p.depth += 4
}

func (p *ncPrinter) dec() {
	p.depth -= 4
}

func (p *ncPrinter) String() string {
	return p.value
}
