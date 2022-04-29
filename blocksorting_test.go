package navigadoc_test

import (
	"strconv"
	"testing"

	"github.com/navigacontentlab/navigadoc"
	"github.com/navigacontentlab/navigadoc/doc"
)

type BlockSortOrderTestCase struct {
	blockSortOrder map[string]int
	blocksToSort   []doc.Block
}

func TestBlockSorter_SortBlocks(t *testing.T) {
	testCases := []BlockSortOrderTestCase{
		{
			blockSortOrder: map[string]int{
				"0": 10,
				"1": 20,
				"2": 30,
				"3": 40,
			},
			blocksToSort: []doc.Block{
				{Type: "3"},
				{Type: "2"},
				{Type: "0"},
				{Type: "1"},
			},
		},
		{
			blockSortOrder: map[string]int{
				"0": 10,
				"1": 20,
				"2": 30,
				"3": 40,
			},
			blocksToSort: []doc.Block{
				{Type: "0"},
				{Type: "1"},
				{Type: "2"},
				{Type: "3"},
			},
		},
		{
			blockSortOrder: map[string]int{
				"0": 0,
				"1": 1,
				"2": 2,
				"3": 2,
			},
			blocksToSort: []doc.Block{
				{Type: "1"},
				{Type: "2"},
				{Type: "0"},
				{Type: "3"},
			},
		},
	}

	for _, tc := range testCases {
		sorter := navigadoc.NewBlockSorter(tc.blockSortOrder)
		sortedBlocks := sorter.SortBlocks(tc.blocksToSort)

		for i, block := range sortedBlocks {
			if block.Type != strconv.Itoa(i) {
				t.Errorf("expected Type=%d, but was %s", i, block.Type)
			}
		}
	}
}
