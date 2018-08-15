package test

import (
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInit(t *testing.T) {
	h := newHarness()
	h.verifyHandlerRegistrations(t)
}

func TestSdkUnknownOperation(t *testing.T) {
	h := newHarness()

	h.expectStateStorageBlockHeightRequested(12)
	h.expectNativeContractMethodCalled("Contract1", "method1", func(contextId primitives.ExecutionContextId) (protocol.ExecutionResult, error) {
		_, err := h.handleSdkCall(contextId, "Sdk.UnknownOperation", "read")
		require.Error(t, err, "handleSdkCall should fail")
		return protocol.EXECUTION_RESULT_SUCCESS, nil
	})

	h.runLocalMethod("Contract1", "method1")

	h.verifyStateStorageBlockHeightRequested(t)
	h.verifyNativeContractMethodCalled(t)
}