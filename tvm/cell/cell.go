package cell

import (
	"encoding/hex"
	"strings"
)

type Cell struct {
	bitsSz int
	index  int
	data   []byte

	refs []*Cell
}

func (c *Cell) BeginParse() *LoadCell {
	// copy data
	data := append([]byte{}, c.data...)

	refs := make([]*LoadCell, len(c.refs))
	for i, ref := range c.refs {
		refs[i] = ref.BeginParse()
	}

	return &LoadCell{
		bitsSz:   c.bitsSz,
		loadedSz: 0,
		data:     data,
		refs:     refs,
	}
}

func (c *Cell) Dump() string {
	return c.dump(0)
}

func (c *Cell) dump(deep int) string {
	str := strings.Repeat("  ", deep) + "[" + hex.EncodeToString(c.data) + "]" + " -> {"
	if len(c.refs) > 0 {
		for _, ref := range c.refs {
			str += "\n" + ref.dump(deep+1) + ", " + "\n"
		}
		str += strings.Repeat("  ", deep)
	}

	return str + "}"
}
