package blockchain

type BlockChain struct {
	Blocks []*Block
}

func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{GenerateGenesisBlock()}}
}

func (b *BlockChain) AddBlock(data string) {

	if len(b.Blocks) == 0 {
		panic("Block chain must have genesis block")
		return
	}

	prevBlock := b.Blocks[len(b.Blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	b.Blocks = append(b.Blocks, newBlock)
}
