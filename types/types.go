package types

type BlockInfo struct {
	BlockHash     string `json:"block_hash"`
	BlockHeight   int    `json:"block_height"`
	Timestamp     int64  `json:"timestamp"`
	TxCount       int    `json:"tx_count"`
	MineByAddress string `json:"mine_by_address"`
	BlockReward   int    `json:"block_reward"`
}

type TransactionInfo struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          int    `json:"amount"`
	Timestamp       int64  `json:"timestamp"`
	BlockHeight     int    `json:"block_height"`
	TransactionHash string `json:"transaction_hash"`
	TransactionFee  int    `json:"transaction_fee"`
}
