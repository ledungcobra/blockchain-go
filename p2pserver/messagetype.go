package p2pserver

type Addr struct {
	AddrList []string
}

type Block struct {
	AddrFrom string
	Block    []byte
}

type GetBlocks struct {
	AddrFrom string
}

type GetData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type Inventory struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type Tx struct {
	AddrFrom string
	Data     []byte
}

type Version struct {
	Version    int
	BestHeight int
	AddrFrom   string
	LastHash   string
}

type BuildBlockChain struct {
	AddrFrom string
}

type SendGetAddr struct {
	AddrFrom string
}
