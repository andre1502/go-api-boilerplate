package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/elastic"
	"go-api-boilerplate/module/logger"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func NewKaiaClient(blockchain *Blockchain) (*KaiaClient, error) {
	if module.IsEmptyString(blockchain.config.KAIA_ENDPOINT) {
		fmt.Println(ErrEmptyKaiaEndpoint)
		logger.Log.Error(ErrEmptyKaiaEndpoint)

		return nil, ErrEmptyKaiaEndpoint
	}

	client, err := ethclient.Dial(blockchain.config.KAIA_ENDPOINT)
	if err != nil {
		msg := "error when connect to the kaia endpoint. %v"
		fmt.Println(fmt.Errorf(msg, err))
		logger.Log.Errorf(msg, err)

		return nil, ErrConnectionKaiaEndpoint
	}

	chainID := big.NewInt(KAIA_MAINNET_CHAIN_ID)
	if blockchain.config.KAIA_TEST_MODE == "True" {
		chainID = big.NewInt(KAIA_KAIROS_TESTNET_CHAIN_ID)
	}

	kaiaClient := &KaiaClient{
		Blockchain: blockchain,
		client:     client,
		chainID:    chainID,
	}

	kaiaClient, err = kaiaClient.InitSender()
	if err != nil {
		fmt.Println(err)
		logger.Log.Error(err)

		return nil, err
	}

	return kaiaClient, nil
}

func (k *KaiaClient) InitSender() (*KaiaClient, error) {
	senderPrivateKey := k.RemoveHexPrefix(k.config.KAIA_SENDER_PRIVATE_KEY)

	if module.IsEmptyString(senderPrivateKey) {
		return k, ErrEmptyKaiaPrivateKey
	}

	privateKey, err := crypto.HexToECDSA(senderPrivateKey)
	if err != nil {
		return k, err
	}

	// Get the public address from the private key
	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return k, ErrKaiaPublicKeyAssert
	}

	k.privateKey = privateKey
	k.publicKey = publicKeyECDSA
	k.fromAddress = crypto.PubkeyToAddress(*publicKeyECDSA)

	return k, nil
}

// weiToKlay converts a *big.Int representing Wei/peb to a human-readable KLAY string
func (k *KaiaClient) weiToKlayString(wei *big.Int, decimals int) string {
	// Create a big.Float from the wei amount
	fWei := new(big.Float).SetInt(wei)

	// Create a big.Float for the divisor (10^decimals)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))

	// Perform the division
	klay := new(big.Float).Quo(fWei, divisor)

	// Format to a string with a reasonable number of decimal places for display
	// Using -1 for precision keeps all digits after the decimal point
	return klay.Text('f', decimals) // 'f' for fixed-point notation, decimals for precision
}

// weiToKlay converts a *big.Int representing Wei/peb to a human-readable KLAY float64
func (k *KaiaClient) weiToKlayFloat64(wei *big.Int, decimals int) float64 {
	// Create a big.Float from the wei amount
	fWei := new(big.Float).SetInt(wei)

	// Create a big.Float for the divisor (10^decimals)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))

	// Perform the division
	klay := new(big.Float).Quo(fWei, divisor)
	klayFloat64, _ := klay.Float64()

	return klayFloat64
}

// weiToGweiString converts a *big.Int representing Wei/peb to Gwei/Gkei string
func (k *KaiaClient) weiToGweiString(wei *big.Int) string {
	// 1 Gwei = 10^9 Wei
	gweiDivisor := big.NewInt(ONE_GWEI)
	gwei := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt(gweiDivisor))
	return gwei.Text('f', 9) // 9 decimal places for Gwei if needed, but often shown as integer
}

// weiToGweiFloat64 converts a *big.Int representing Wei/peb to Gwei/Gkei float64
func (k *KaiaClient) weiToGweiFloat64(wei *big.Int) float64 {
	// 1 Gwei = 10^9 Wei
	gweiDivisor := big.NewInt(ONE_GWEI)
	gwei := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt(gweiDivisor))
	gweiFloat64, _ := gwei.Float64()

	return gweiFloat64
}

// Convert the amount from KLAY to Wei
func (k *KaiaClient) klayToWei(amountKlay float64) *big.Int {
	// Convert KLAY to Wei (1 KLAY = 10^18 Wei)
	// Multiplying by 1e6 then 1e12 to handle float precision safely for smaller numbers
	return new(big.Int).Mul(big.NewInt(int64(amountKlay*1000000)), big.NewInt(1e12))
}

func (k *KaiaClient) getTxTypeName(txType uint8) string {
	switch txType {
	case types.LegacyTxType:
		return TX_TYPE_LEGACY
	case types.AccessListTxType:
		return TX_TYPE_EIP_2930
	case types.DynamicFeeTxType:
		return TX_TYPE_ETHEREUM_DYNAMIC_FEE
	default:
		return fmt.Sprintf("Unknown Type (%d)", txType)
	}
}

func (k *KaiaClient) SendTransaction(ctx context.Context, amountKlay float64, recipientAddress string) (*KaiaSendTransaction, error) {
	if k.privateKey == nil {
		return nil, ErrEmptyKaiaPrivateKey
	}

	if amountKlay == 0 {
		return nil, ErrKlayAmountZero
	}

	if module.IsEmptyString(recipientAddress) {
		return nil, ErrRecipientAddressEmpty
	}

	// Get the current nonce (transaction count) for the sender's address
	nonce, err := k.client.PendingNonceAt(ctx, k.fromAddress)
	if err != nil {
		logger.Log.Errorf("error when get current nonce for transaction: %v", err)
		return nil, ErrSenderAddressPendingNonce
	}

	amountWei := k.klayToWei(amountKlay)

	// Define the recipient address
	toAddress := common.HexToAddress(recipientAddress)

	// Gas limit for a simple KLAY transfer is typically 21000.
	// Using 25000 as a safe buffer for simple transfers.
	gasLimit := uint64(25000)

	// Get the latest block header to fetch BaseFeePerGas
	header, err := k.client.HeaderByNumber(ctx, nil) // nil for latest block
	if err != nil {
		logger.Log.Errorf("error when get latest block header to fetch BaseFeePerGas: %v", err)
		return nil, ErrGetLatestBlockHeader
	}

	var txData types.TxData

	if header.BaseFee == nil {
		// network does not support EIP-1559 (BaseFeePerGas is nil). Please use LegacyTx

		// Get the suggested gas price
		gasPrice, err := k.client.SuggestGasPrice(ctx)
		if err != nil {
			logger.Log.Errorf("error when get suggested gas price: %v", err)
			return nil, ErrGetSuggestedGasPrice
		}

		txData = &types.LegacyTx{
			Nonce:    nonce,
			To:       &toAddress, // Use pointer for common.Address
			Value:    amountWei,
			Gas:      gasLimit,
			GasPrice: gasPrice,
			Data:     nil, // Data is nil for simple transfers
		}
	} else {
		baseFeePerGas := header.BaseFee

		// Define MaxPriorityFeePerGas (tip to miner)
		// You can adjust this based on network congestion or desired transaction speed.
		// 1 Gwei = 1,000,000,000 Wei
		maxPriorityFeePerGas := big.NewInt(ONE_GWEI) // 1 Gwei

		// Calculate MaxFeePerGas
		// A common heuristic: 2 * BaseFeePerGas + MaxPriorityFeePerGas
		// This ensures you pay enough even if base fee fluctuates, and get a refund if you overbid.
		maxFeePerGas := new(big.Int).Add(
			new(big.Int).Mul(baseFeePerGas, big.NewInt(2)),
			maxPriorityFeePerGas,
		)

		// Create a new DynamicFee transaction (Type 2) using types.NewTx
		txData = &types.DynamicFeeTx{
			ChainID:   k.chainID,
			Nonce:     nonce,
			GasTipCap: maxPriorityFeePerGas,
			GasFeeCap: maxFeePerGas,
			Gas:       gasLimit,
			To:        &toAddress, // Use pointer for common.Address
			Value:     amountWei,
			Data:      nil, // Data is nil for simple transfers
			// AccessList: nil, // Optional: EIP-2930 access list
		}
	}

	tx := types.NewTx(txData)

	// Sign the transaction
	// Use LatestSignerForChainID for EIP-1559 compatibility
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(k.chainID), k.privateKey)
	if err != nil {
		logger.Log.Errorf("failed to sign transaction: %v", err)
		return nil, ErrSignTransaction
	}

	// Send the signed transaction to the network
	err = k.client.SendTransaction(ctx, signedTx)
	if err != nil {
		logger.Log.Errorf("error when send signed transaction: %v", err)
		return nil, ErrSendSignedTransaction
	}

	// Wait for the transaction to be mined and get the receipt
	receipt, err := bind.WaitMined(ctx, k.client, signedTx.Hash())
	if err != nil {
		logger.Log.Errorf("error when waiting for mined transaction receipt: %v", err)
		return nil, ErrMinedTransactionReceipt
	}

	transactionID := signedTx.Hash().Hex()
	transactionDate := signedTx.Time()
	blockNumber := receipt.BlockNumber.Uint64()
	gasUsed := receipt.GasUsed
	receiptStatus := receipt.Status

	logger.Log.WithFields(map[string]interface{}{
		"elastic_index":    elastic.ELASTIC_TRANSACTION_ACTIVITY_INDEX,
		"from_address":     k.fromAddress,
		"to_address":       recipientAddress,
		"transaction_id":   transactionID,
		"transaction_date": transactionDate,
		"block_number":     blockNumber,
		"gas_used":         gasUsed,
		"receipt_status":   receipt.Status,
	}).Infof("send transaction from %s to %s.", k.fromAddress, recipientAddress)

	return &KaiaSendTransaction{
		TransactionID:   transactionID,
		TransactionDate: transactionDate,
		BlockNumber:     blockNumber,
		GasUsed:         gasUsed,
		ReceiptStatus:   receiptStatus,
	}, nil
}

func (k *KaiaClient) SearchTransactionHash(ctx context.Context, fromAddress, toAddress string, transactionHash string) (*KaiaSearchTransaction, error) {
	sourceAddress := common.HexToAddress(fromAddress)
	targetAddress := common.HexToAddress(toAddress)
	txHash := common.HexToHash(transactionHash)

	// Fetch the transaction by hash
	tx, isPending, err := k.client.TransactionByHash(ctx, txHash)
	if err != nil {
		logger.Log.Errorf("error when get transaction from hash: %v", err)
		return nil, ErrGetTransactionFromHash
	}

	if tx == nil {
		return nil, ErrTransactionNotFound
	}

	// Get the transaction receipt
	// This is essential for details like GasUsed, EffectiveGasPrice, Status, and Logs
	receipt, err := k.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		logger.Log.Errorf("error when get transaction receipt: %v", err)
		return nil, ErrGetTransactionReceipt
	}

	var blockNumber *big.Int
	var blockBaseFee *big.Int
	var txFeeWei *big.Int
	var gasUsed *uint64
	var gasUsagePercentage float64
	var burntFee *big.Int
	var gasPrice *big.Int
	var transactionDate *time.Time
	var receiptStatus *uint64

	if receipt != nil {
		if receipt.BlockNumber != nil {
			block, err := k.client.HeaderByNumber(ctx, receipt.BlockNumber)
			if err != nil {
				logger.Log.Errorf("error when get block header: %v", err)
				return nil, ErrGetBlockHeader
			}

			blockNumber = receipt.BlockNumber

			if block.BaseFee != nil {
				blockBaseFee = block.BaseFee
			}

			txDate := time.Unix(int64(block.Time), 0)
			transactionDate = &txDate
		}

		// Gas Limit & Usage by Tx
		gasUsed = &receipt.GasUsed
		gasUsagePercentage = float64(receipt.GasUsed) / float64(tx.Gas()) * 100

		// Burnt Fee (EIP-1559 only)
		if receipt.EffectiveGasPrice != nil {
			gasPrice = receipt.EffectiveGasPrice
			effectiveGasPriceFloat := new(big.Float).SetInt(receipt.EffectiveGasPrice)

			// Divide Effective Gas Price by 2 (as observed in the image for the burnt fee ratio)
			halfEffectiveGasPriceFloat := new(big.Float).Quo(effectiveGasPriceFloat, big.NewFloat(2.0))
			burntFeeFloat := new(big.Float).Mul(halfEffectiveGasPriceFloat, new(big.Float).SetUint64(receipt.GasUsed))

			burntFee = new(big.Int)
			burntFeeFloat.Int(burntFee) // Convert to big.Int, truncating decimals
		} else if blockBaseFee != nil {
			burntFee = new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), blockBaseFee)
		} else {
			// Burnt Fee: 0 KAIA (Not an EIP-1559 transaction or BaseFee not available in block header)
			burntFee = big.NewInt(0)
		}

		receiptStatus = &receipt.Status

		// TX Fee
		// Total TX Fee = Gas Used * Gas Price
		txFeeWei = new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), gasPrice)
	}

	// Gas Price (from transaction object)
	// This is tx.GasPrice() for Type 0, or tx.GasFeeCap() for Type 2
	if gasPrice == nil {
		if tx.Type() == types.DynamicFeeTxType {
			gasPrice = tx.GasTipCap() // Max fee user was willing to pay
		} else {
			gasPrice = tx.GasPrice() // Price for legacy tx
		}
	}

	// Recover the sender address
	senderAddress, err := types.Sender(types.LatestSignerForChainID(k.chainID), tx)
	if err != nil {
		logger.Log.Errorf("error when recover sender address from transaction: %v", err)
		return nil, ErrRecoverSenderAddressFromTransaction
	}

	// Get the recipient address (will be nil for contract creation transactions)
	var recipientAddress common.Address
	if tx.To() != nil {
		recipientAddress = *tx.To()
	} else {
		// Tx Recipient: Contract creation transaction (no direct 'to' address)
		// For contract creation, the 'to' address is the address of the newly deployed contract.
		// You'd typically get this from the transaction receipt if the transaction was successful.
	}

	kaiaSearchTransaction := &KaiaSearchTransaction{
		SenderAddress:    senderAddress.Hex(),
		RecipientAddress: recipientAddress.Hex(),
		TransactionHash:  txHash.Hex(),
		BlockNumber:      blockNumber.String(),
		TransactionType:  k.getTxTypeName(tx.Type()),
		TransactionDate:  transactionDate,
		TransactionFee:   k.weiToKlayFloat64(txFeeWei, KLAYTN_DECIMALS),
		IsPending:        isPending,
		ReceiptStatus:    receiptStatus,
		Amount:           k.weiToKlayFloat64(tx.Value(), KLAYTN_DECIMALS),
		Gas:              tx.Gas(),
		GasUsed:          gasUsed,
		GasUsageRate:     gasUsagePercentage,
		GasPrice:         k.weiToKlayFloat64(gasPrice, KLAYTN_DECIMALS),
		BurntFee:         k.weiToKlayFloat64(burntFee, KLAYTN_DECIMALS),
		Nonce:            tx.Nonce(),
		Data:             common.Bytes2Hex(tx.Data()),
	}

	if sourceAddress == senderAddress {
		kaiaSearchTransaction.IsSourceSameAsSenderAddress = true
	}

	if targetAddress == recipientAddress {
		kaiaSearchTransaction.IsTargetSameAsRecipientAddress = true
	}

	kaiaSearchTransaction.IsRelatedTransaction = kaiaSearchTransaction.IsSourceSameAsSenderAddress && kaiaSearchTransaction.IsTargetSameAsRecipientAddress

	return kaiaSearchTransaction, nil
}
