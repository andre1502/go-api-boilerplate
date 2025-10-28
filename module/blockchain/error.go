package blockchain

import "errors"

var (
	ErrEmptyKaiaEndpoint                   = errors.New("kaia endpoint is empty, please check .env is setup correctly")
	ErrConnectionKaiaEndpoint              = errors.New("error when connect to the kaia endpoint")
	ErrEmptyKaiaPrivateKey                 = errors.New("kaia private key is empty, please check .env is setup correctly")
	ErrKaiaPublicKeyAssert                 = errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	ErrKlayAmountZero                      = errors.New("klay amount is 0")
	ErrRecipientAddressEmpty               = errors.New("empty recipient address")
	ErrTransactionNotFound                 = errors.New("transaction not found")
	ErrSenderAddressPendingNonce           = errors.New("error when get current nonce for sender address")
	ErrRecoverSenderAddressFromTransaction = errors.New("error when recover sender address from transaction")
	ErrGetLatestBlockHeader                = errors.New("error when get latest block header to fetch BaseFeePerGas")
	ErrGetBlockHeader                      = errors.New("error when get block header")
	ErrGetSuggestedGasPrice                = errors.New("error when get suggested gas price")
	ErrSignTransaction                     = errors.New("error when sign transaction")
	ErrSendSignedTransaction               = errors.New("error when send signed transaction")
	ErrMinedTransactionReceipt             = errors.New("error when waiting for mined transaction receipt")
	ErrGetTransactionFromHash              = errors.New("error when get transaction from hash")
	ErrGetTransactionReceipt               = errors.New("error when get transaction receipt")
)
