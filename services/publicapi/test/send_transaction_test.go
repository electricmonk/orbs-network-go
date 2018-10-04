package test

import (
	"context"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/orbs-network-go/crypto/digest"
	"github.com/orbs-network/orbs-network-go/test"
	"github.com/orbs-network/orbs-network-go/test/builders"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/orbs-network/orbs-spec/types/go/services/handlers"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSendTransaction_CallsTxPool(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		harness := newPublicApiHarness(ctx, 1*time.Millisecond)

		harness.txpMock.When("AddNewTransaction", mock.Any).Return(&services.AddNewTransactionOutput{}).Times(1)

		harness.papi.SendTransaction(&services.SendTransactionInput{
			ClientRequest: (&client.SendTransactionRequestBuilder{
				SignedTransaction: builders.Transaction().Builder()}).Build(),
		})

		ok, err := harness.txpMock.Verify()
		require.True(t, ok, "should have called the txp func")
		require.NoError(t, err, "error happened when it should not")
	})
}

func TestSendTransaction_AlreadyCommitted(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		harness := newPublicApiHarness(ctx, 1*time.Millisecond)

		harness.txpMock.When("AddNewTransaction", mock.Any).Return(&services.AddNewTransactionOutput{
			TransactionStatus:  protocol.TRANSACTION_STATUS_DUPLICATE_TRANSACTION_ALREADY_COMMITTED,
			TransactionReceipt: builders.TransactionReceipt().Build(),
		}).Times(1)

		result, err := harness.papi.SendTransaction(&services.SendTransactionInput{
			ClientRequest: (&client.SendTransactionRequestBuilder{
				SignedTransaction: builders.Transaction().Builder()}).Build(),
		})

		require.NoError(t, err, "error happened when it should not")
		require.NotNil(t, result, "Send transaction returned nil instead of object")
	})
}

func TestSendTransaction_BlocksUntilTransactionCompletes(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		harness := newPublicApiHarness(ctx, 1*time.Second)

		txb := builders.Transaction().Builder()
		harness.onAddNewTransaction(func() {
			harness.papi.HandleTransactionResults(&handlers.HandleTransactionResultsInput{
				TransactionReceipts: []*protocol.TransactionReceipt{builders.TransactionReceipt().WithTransaction(txb.Build().Transaction()).Build()},
			})
		})

		result, err := harness.papi.SendTransaction(&services.SendTransactionInput{
			ClientRequest: (&client.SendTransactionRequestBuilder{
				SignedTransaction: txb,
			}).Build(),
		})

		require.NoError(t, err, "error happened when it should not")
		require.NotNil(t, result, "Send transaction returned nil instead of object")
	})
}

func TestSendTransaction_BlocksUntilTransactionErrors(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		harness := newPublicApiHarness(ctx, 1*time.Second)

		txb := builders.Transaction().Builder()
		txHash := digest.CalcTxHash(txb.Build().Transaction())

		harness.onAddNewTransaction(func() {
			harness.papi.HandleTransactionError(&handlers.HandleTransactionErrorInput{
				Txhash:            txHash,
				TransactionStatus: protocol.TRANSACTION_STATUS_REJECTED_TIMESTAMP_WINDOW_EXCEEDED,
			})
		})

		result, err := harness.papi.SendTransaction(&services.SendTransactionInput{
			ClientRequest: (&client.SendTransactionRequestBuilder{
				SignedTransaction: txb,
			}).Build(),
		})

		require.NoError(t, err, "error happened when it should not")
		require.NotNil(t, result, "Send transaction returned nil instead of object")
		require.Equal(t, protocol.TRANSACTION_STATUS_REJECTED_TIMESTAMP_WINDOW_EXCEEDED, result.ClientResponse.TransactionStatus(), "got wrong status")
	})
}