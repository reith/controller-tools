package deepcopy

import (
	"fmt"
	"io"
)

// NB(directxman12): This code is a bit of a byzantine mess.
// I've tried to clean it up a bit from the original in deepcopy-gen,
// but parts remain a bit convoluted.  Exercise caution when changing.
// It's perhaps a tad over-commented now, but better safe than sorry.
// It also seriously needs auditing for sanity -- there's parts where we
// copy the original deepcopy-gen's output just to be safe, but some of that
// could be simplified away if we're careful.

// CodeWriter assists in writing out Go code lines and blocks to a writer.
type CodeWriter struct {
	out io.Writer
}

// Line writes a single line.
func (c *CodeWriter) Line(line string) {
	fmt.Fprintln(c.out, line)
}

// Linef writes a single line with formatting (as per fmt.Sprintf).
func (c *CodeWriter) Linef(line string, args ...interface{}) {
	fmt.Fprintf(c.out, line+"\n", args...)
}

// If writes an if statement with the given setup/condition clause, executing
// the given function to write the contents of the block.
func (c *CodeWriter) If(setup string, block func()) {
	c.Linef("if %s {", setup)
	block()
	c.Line("}")
}

// If writes if and else statements with the given setup/condition clause, executing
// the given functions to write the contents of the blocks.
func (c *CodeWriter) IfElse(setup string, ifBlock func(), elseBlock func()) {
	c.Linef("if %s {", setup)
	ifBlock()
	c.Line("} else {")
	elseBlock()
	c.Line("}")
}

// For writes an for statement with the given setup/condition clause, executing
// the given function to write the contents of the block.
func (c *CodeWriter) For(setup string, block func()) {
	c.Linef("for %s {", setup)
	block()
	c.Line("}")
}

func NewCodeWriter(out io.Writer) *CodeWriter {
	return &CodeWriter{out: out}
}
