package navigadoc

import "github.com/navigacontentlab/navigadoc/doc"

type BlockReplacement interface {
	GetOldBlock() doc.Block
	GetNewBlock() doc.Block
}

func MergeProperties(existingProperties []doc.Property, newProperties []doc.Property) []doc.Property {
	resultDict := make(map[string]doc.Property)
	for _, existingProperty := range existingProperties {
		resultDict[existingProperty.Name] = existingProperty
	}
	for _, newProperty := range newProperties {
		resultDict[newProperty.Name] = newProperty
	}
	resultProperties := make([]doc.Property, 0, len(resultDict))

	for _, v := range resultDict {
		resultProperties = append(resultProperties, v)
	}
	return resultProperties
}

func DeleteBlocks(document doc.Document, blocksToDelete []doc.Block) *doc.Document {
	// process links
	document.Links = GetBlocksToKeep(document.Links, blocksToDelete)

	// process meta
	document.Meta = GetBlocksToKeep(document.Meta, blocksToDelete)

	// process content
	document.Content = GetBlocksToKeep(document.Content, blocksToDelete)
	return &document
}

//goland:noinspection GoUnusedExportedFunction
func GetBlocks(document doc.Document, patterns []doc.Block) []doc.Block {
	var foundBlocks []doc.Block
	foundBlocks = append(foundBlocks, getBlocks(document.Meta, patterns)...)
	foundBlocks = append(foundBlocks, getBlocks(document.Links, patterns)...)
	foundBlocks = append(foundBlocks, getBlocks(document.Content, patterns)...)
	return foundBlocks
}

func getBlocks(blocks []doc.Block, patterns []doc.Block) []doc.Block {
	var foundBlocks []doc.Block
	for _, block := range blocks {
		for _, pattern := range patterns {
			if matchBlock(pattern, block) {
				foundBlocks = append(foundBlocks, block)
				// remember the block to delete
				break
			}
		}

		block.Links = getBlocks(block.Links, patterns)
		block.Meta = getBlocks(block.Meta, patterns)
		block.Content = getBlocks(block.Content, patterns)
	}
	return foundBlocks
}

// returns block NOT matching the pattern
func GetBlocksToKeep(blocks []doc.Block, patterns []doc.Block) []doc.Block {
	// first get the blocks to delete
	var blocksToKeep []doc.Block
	for _, block := range blocks {
		shouldBeDeleted := false
		for _, pattern := range patterns {
			if matchBlock(pattern, block) {
				shouldBeDeleted = true
				// remember the block to delete
				break
			}
		}

		block.Links = GetBlocksToKeep(block.Links, patterns)
		block.Meta = GetBlocksToKeep(block.Meta, patterns)
		block.Content = GetBlocksToKeep(block.Content, patterns)

		if !shouldBeDeleted {
			blocksToKeep = append(blocksToKeep, block)
		}
	}

	return blocksToKeep
}

func matchBlock(pattern, block doc.Block) bool {
	// Id
	if pattern.ID != "" && pattern.ID != block.ID {
		return false
	}

	// UUID
	if pattern.UUID != "" && pattern.UUID != block.UUID {
		return false
	}

	// URI
	if pattern.URI != "" && pattern.URI != block.URI {
		return false
	}

	// URL
	if pattern.URL != "" && pattern.URL != block.URL {
		return false
	}

	// Type
	if pattern.Type != "" && pattern.Type != block.Type {
		return false
	}

	// Title
	if pattern.Title != "" && pattern.Title != block.Title {
		return false
	}

	// Rel
	if pattern.Rel != "" && pattern.Rel != block.Rel {
		return false
	}

	// Name
	if pattern.Name != "" && pattern.Name != block.Name {
		return false
	}

	// Value
	if pattern.Value != "" && pattern.Value != block.Value {
		return false
	}

	// ContentType
	if pattern.ContentType != "" && pattern.ContentType != block.ContentType {
		return false
	}

	// Role
	if pattern.Role != "" && pattern.Role != block.Role {
		return false
	}

	return true
}

func DeDuplicateLinks(links []doc.Block, linksToDeDuplicate []doc.Block) []doc.Block {
	var isDuplicate bool
	var uniqueList []doc.Block
	for _, link := range links {
		isDuplicate = false
		for _, p := range linksToDeDuplicate {
			// check if link should be unique
			if matchBlock(p, link) {
				// see if already in list
				for _, u := range uniqueList {
					if matchBlock(u, link) {
						isDuplicate = true
					}
				}
			}
		}
		if !isDuplicate {
			uniqueList = append(uniqueList, link)
		}
	}
	return uniqueList
}

func ReplaceBlocks(document doc.Document, blocksToReplace []BlockReplacement) *doc.Document {
	document.Links = replaceBlocksInList(blocksToReplace, document.Links)
	document.Meta = replaceBlocksInList(blocksToReplace, document.Meta)
	document.Content = replaceBlocksInList(blocksToReplace, document.Content)
	return &document
}

func replaceBlocksInList(blocksToReplace []BlockReplacement, blocks []doc.Block) []doc.Block {
	var newBlocks []doc.Block
	for _, block := range blocks {
		for _, blockToReplace := range blocksToReplace {
			if matchBlock(blockToReplace.GetOldBlock(), block) {
				block = replaceBlock(blockToReplace.GetNewBlock(), block)
			}
		}
		block.Links = replaceBlocksInList(blocksToReplace, block.Links)
		block.Meta = replaceBlocksInList(blocksToReplace, block.Meta)
		block.Content = replaceBlocksInList(blocksToReplace, block.Content)
		newBlocks = append(newBlocks, block)
	}
	return newBlocks
}

func replaceBlock(newBlock, block doc.Block) doc.Block {
	if newBlock.ID != "" {
		block.ID = newBlock.ID
	}
	if newBlock.UUID != "" {
		block.UUID = newBlock.UUID
	}
	if newBlock.URI != "" {
		block.URI = newBlock.URI
	}
	if newBlock.URL != "" {
		block.URL = newBlock.URL
	}
	if newBlock.Type != "" {
		block.Type = newBlock.Type
	}
	if newBlock.Title != "" {
		block.Title = newBlock.Title
	}
	if newBlock.Rel != "" {
		block.Rel = newBlock.Rel
	}
	if newBlock.Name != "" {
		block.Name = newBlock.Name
	}
	if newBlock.Value != "" {
		block.Value = newBlock.Value
	}
	if newBlock.ContentType != "" {
		block.ContentType = newBlock.ContentType
	}
	if newBlock.Role != "" {
		block.Role = newBlock.Role
	}
	if len(newBlock.Data) > 0 {
		block.Data = newBlock.Data
	}
	return block
}

func FilterBlocks(blocks []doc.Block, filter func(a doc.Block) bool) []doc.Block {
	var filteredBlocks []doc.Block
	for _, currentBlock := range blocks {
		if filter(currentBlock) {
			filteredBlocks = append(filteredBlocks, currentBlock)
		}
	}
	return filteredBlocks
}
