package blockchain

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type KaiaClient struct {
	*Blockchain
	client      *ethclient.Client
	privateKey  *ecdsa.PrivateKey
	publicKey   *ecdsa.PublicKey
	fromAddress common.Address
	chainID     *big.Int
}

type KaiaSendTransaction struct {
	TransactionID   string    `json:"transaction_id"`
	TransactionDate time.Time `json:"transaction_date"`
	BlockNumber     uint64    `json:"block_number"`
	GasUsed         uint64    `json:"gas_used"`
	ReceiptStatus   uint64    `json:"receipt_status"`
}

type KaiaSearchTransaction struct {
	SenderAddress                  string     `json:"sender_address"`
	RecipientAddress               string     `json:"receipient_address"`
	IsSourceSameAsSenderAddress    bool       `json:"is_source_same_as_sender_address"`
	IsTargetSameAsRecipientAddress bool       `json:"is_target_same_as_recipient_address"`
	IsRelatedTransaction           bool       `json:"is_related_transaction"`
	TransactionHash                string     `json:"transaction_hash"`
	BlockNumber                    string     `json:"block_number"`
	TransactionType                string     `json:"transaction_type"`
	TransactionDate                *time.Time `json:"transaction_date"`
	TransactionFee                 float64    `json:"transaction_fee"`
	IsPending                      bool       `json:"is_pending"`
	ReceiptStatus                  *uint64    `json:"receipt_status"`
	Amount                         float64    `json:"amount"`
	Gas                            uint64     `json:"gas"`
	GasUsed                        *uint64    `json:"gas_used"`
	GasUsageRate                   float64    `json:"gas_usage_rate"`
	GasPrice                       float64    `json:"gas_price"`
	BurntFee                       float64    `json:"burnt_fee"`
	Nonce                          uint64     `json:"nonce"`
	Data                           string     `json:"data"`
}
