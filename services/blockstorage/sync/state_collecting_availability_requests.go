package sync

import (
	"context"
	"github.com/orbs-network/orbs-network-go/instrumentation/log"
	"github.com/orbs-network/orbs-network-go/synchronization"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol/gossipmessages"
	"github.com/orbs-network/orbs-spec/types/go/services/gossiptopics"
)

type collectingAvailabilityResponsesState struct {
	sf        *stateFactory
	gossip    gossiptopics.BlockSync
	storage   BlockSyncStorage
	config    blockSyncConfig
	responses []*gossipmessages.BlockAvailabilityResponseMessage
	logger    log.BasicLogger
}

func (s *collectingAvailabilityResponsesState) name() string {
	return "collecting-availability-responses"
}

func (s *collectingAvailabilityResponsesState) processState(ctx context.Context) syncState {
	s.responses = []*gossipmessages.BlockAvailabilityResponseMessage{}

	err := s.petitionerBroadcastBlockAvailabilityRequest()
	if err != nil {
		s.logger.Info("failed to broadcast block availability request", log.Error(err))
		return s.sf.CreateIdleState()
	}

	waitForResponses := synchronization.NewTimer(s.config.BlockSyncCollectResponseTimeout())
	select {
	case <-waitForResponses.C:
		return s.sf.CreateFinishedCARState(s.responses)
	case <-ctx.Done():
		return nil
	}
}

func (s *collectingAvailabilityResponsesState) blockCommitted(blockHeight primitives.BlockHeight) {
	return
}

func (s *collectingAvailabilityResponsesState) gotAvailabilityResponse(message *gossipmessages.BlockAvailabilityResponseMessage) {
	s.responses = append(s.responses, message)
}

func (s *collectingAvailabilityResponsesState) gotBlocks(message *gossipmessages.BlockSyncResponseMessage) {
	return
}

func (s *collectingAvailabilityResponsesState) petitionerBroadcastBlockAvailabilityRequest() error {
	lastCommittedBlockHeight := s.storage.LastCommittedBlockHeight()
	firstBlockHeight := lastCommittedBlockHeight + 1
	lastBlockHeight := lastCommittedBlockHeight + primitives.BlockHeight(s.config.BlockSyncBatchSize())

	s.logger.Info("broadcast block availability request",
		log.Stringable("first-block-height", firstBlockHeight),
		log.Stringable("last-block-height", lastBlockHeight))

	input := &gossiptopics.BlockAvailabilityRequestInput{
		Message: &gossipmessages.BlockAvailabilityRequestMessage{
			Sender: (&gossipmessages.SenderSignatureBuilder{
				SenderPublicKey: s.config.NodePublicKey(),
			}).Build(),
			SignedBatchRange: (&gossipmessages.BlockSyncRangeBuilder{
				BlockType:                gossipmessages.BLOCK_TYPE_BLOCK_PAIR,
				LastBlockHeight:          lastBlockHeight,
				FirstBlockHeight:         firstBlockHeight,
				LastCommittedBlockHeight: lastCommittedBlockHeight,
			}).Build(),
		},
	}

	_, err := s.gossip.BroadcastBlockAvailabilityRequest(input)
	return err
}
