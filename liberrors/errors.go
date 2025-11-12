package liberrors

import "errors"

var (
	ErrGetBlock         = errors.New("failed to get block")
	ErrUnsupported      = errors.New("unsupported")
	ErrAbiError         = errors.New("failed to parse ABI")
	ErrRPCClient        = errors.New("rpc client not working")
	ErrFailedToGetNonce = errors.New("failed to get nonce")

	ErrInvalidPrivateKey     = errors.New("invalid private key")
	ErrFailedToGetBalance    = errors.New("failed to get balance")
	ErrFailedToSignTx        = errors.New("failed to sign tx")
	ErrFailedToSendTx        = errors.New("failed to send tx")
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrInsufficientAllowance = errors.New("insufficient allowance")
	ErrGasEstimation         = errors.New("failed to estimate gas cost")
)
