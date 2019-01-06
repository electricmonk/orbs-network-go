package consensuscontext

import "errors"

var ErrMismatchedProtocolVersion = errors.New("mismatched protocol version")
var ErrMismatchedVirtualChainID = errors.New("mismatched virtual chain ID")
var ErrMismatchedBlockHeight = errors.New("mismatched block height")
var ErrMismatchedPrevBlockHash = errors.New("mismatched previous block hash")
var ErrInvalidBlockTimestamp = errors.New("invalid current block timestamp")

var ErrIncorrectTransactionOrdering = errors.New("incorrect transaction ordering")

var ErrMismatchedTxRxBlockHeight = errors.New("mismatched block height between transactions and results")
var ErrMismatchedTxRxTimestamps = errors.New("mismatched timestamp between transactions and results")
var ErrMismatchedTxHashPtrToActualTxBlock = errors.New("mismatched tx block hash ptr to actual tx block hash")

var ErrGetStateHash = errors.New("failed in GetStateHash() so cannot retrieve pre-execution state diff merkleRoot from previous block")
var ErrMismatchedPreExecutionStateMerkleRoot = errors.New("pre-execution state diff merkleRoot is different between results block header and extracted from state storage for previous block")
var ErrProcessTransactionSet = errors.New("failed in ProcessTransactionSet()")
