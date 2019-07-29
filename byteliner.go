package byteline

// ByteLiner supports mapping between byte offset and line/column.
// Lines start at 0, columns start at 0.
// The newline character that ends a line is the last column
// on the line it ends.
type ByteLiner interface {
	GetLineAndColumn(byteOffset int) (line int, col int, ok error)
	GetOffset(line int, column int) (offset int, ok error)
	GetCurrentLineAndColumn() (line int, col int, ok error)
	GetCurrentOffset() (offset int, ok error)
}
