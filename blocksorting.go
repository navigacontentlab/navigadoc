package navigadoc

import (
	"sort"

	"github.com/navigacontentlab/navigadoc/doc"
)

type BlockSorter struct {
	Blocks         []doc.Block
	BlockSortOrder map[string]int
}

func NewBlockSorter(blockSortOrder map[string]int) BlockSorter {
	return BlockSorter{BlockSortOrder: blockSortOrder}
}

func (b BlockSorter) Len() int {
	return len(b.Blocks)
}

func (b BlockSorter) Less(i, j int) bool {
	if b.BlockSortOrder[b.Blocks[i].Type] == b.BlockSortOrder[b.Blocks[j].Type] {
		return false
	}

	return b.BlockSortOrder[b.Blocks[i].Type] < b.BlockSortOrder[b.Blocks[j].Type]
}

func (b BlockSorter) Swap(i, j int) { b.Blocks[i], b.Blocks[j] = b.Blocks[j], b.Blocks[i] }

func (b *BlockSorter) SortBlocks(blocks []doc.Block) []doc.Block {
	b.Blocks = blocks
	sort.Stable(b)
	return b.Blocks
}
