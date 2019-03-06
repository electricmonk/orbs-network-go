package virtualmachine

import (
	"github.com/orbs-network/orbs-network-go/crypto/digest"
	"github.com/orbs-network/orbs-network-go/crypto/signature"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
)

func (s *service) verifyTransactionSignatures(signedTransactions []*protocol.SignedTransaction, resultStatuses []protocol.TransactionStatus) {
	for i, signedTransaction := range signedTransactions {

		// skip transactions that already failed due to different reasons
		if resultStatuses[i] != protocol.TRANSACTION_STATUS_RESERVED {
			continue
		}

		// check transaction signature
		switch signedTransaction.Transaction().Signer().Scheme() {
		case protocol.SIGNER_SCHEME_EDDSA:
			if verifyEd25519Signer(signedTransaction) {
				resultStatuses[i] = protocol.TRANSACTION_STATUS_PRE_ORDER_VALID
			} else {
				resultStatuses[i] = protocol.TRANSACTION_STATUS_REJECTED_SIGNATURE_MISMATCH
			}
		default:
			resultStatuses[i] = protocol.TRANSACTION_STATUS_REJECTED_UNKNOWN_SIGNER_SCHEME
		}

	}
}

func verifyEd25519Signer(signedTransaction *protocol.SignedTransaction) bool {
	signerPublicKey := signedTransaction.Transaction().Signer().Eddsa().SignerPublicKey()
	txHash := digest.CalcTxHash(signedTransaction.Transaction())
	return signature.VerifyEd25519(signerPublicKey, txHash, signedTransaction.Signature())
}
