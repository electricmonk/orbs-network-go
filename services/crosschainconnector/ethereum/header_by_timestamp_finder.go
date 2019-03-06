package ethereum

import (
	"context"
	"fmt"
	"github.com/orbs-network/orbs-network-go/instrumentation/log"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/pkg/errors"
	"math"
	"math/big"
	"time"
)

const NEAR_FUTURE_GRACE = 2 * time.Minute

type TimestampFetcher interface {
	GetBlockByTimestamp(ctx context.Context, nano primitives.TimestampNano) (*big.Int, error)
}

type finder struct {
	logger log.BasicLogger
	btg    BlockAndTimestampGetter
}

func NewTimestampFetcher(btg BlockAndTimestampGetter, logger log.BasicLogger) *finder {
	f := &finder{
		btg:    btg,
		logger: logger,
	}

	return f
}

func (f *finder) GetBlockByTimestamp(ctx context.Context, nano primitives.TimestampNano) (*big.Int, error) {
	timestampInSeconds := int64(nano) / int64(time.Second)
	// ethereum started around 2015/07/31
	if timestampInSeconds < 1438300800 {
		return nil, errors.New("cannot query before ethereum genesis")
	}

	latest, err := f.btg.ApproximateBlockAt(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest block")
	}

	if latest == nil { // simulator always returns nil block number
		return nil, nil
	}

	latestNano := uint64(latest.TimeSeconds * int64(time.Second))
	requestedNano := uint64(nano)
	if latestNano < requestedNano && requestedNano-latestNano <= uint64(NEAR_FUTURE_GRACE) {
		return big.NewInt(latest.Number), nil
	}

	// this was added to support simulations and tests, should not be relevant for production
	latestNum := latest.Number
	latestNum -= 10000
	if latestNum < 0 {
		latestNum = 0
	}
	back10k, err := f.btg.ApproximateBlockAt(ctx, big.NewInt(latestNum))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get past reference block")
	}

	theBlock, err := f.findBlockByTimeStamp(ctx, timestampInSeconds, back10k.Number, back10k.TimeSeconds, latest.Number, latest.TimeSeconds)
	return theBlock, err
}

func (f *finder) findBlockByTimeStamp(ctx context.Context, timestampSeconds int64, currentBlockNumber, currentTimestampSeconds, prevBlockNumber, prevTimestampSeconds int64) (*big.Int, error) {
	f.logger.Info("searching for block in ethereum",
		log.Int64("target-timestamp", timestampSeconds),
		log.Int64("current-block-number", currentBlockNumber),
		log.Int64("current-timestamp", currentTimestampSeconds),
		log.Int64("prev-block-number", prevBlockNumber),
		log.Int64("prev-timestamp", prevTimestampSeconds))
	blockNumberDiff := currentBlockNumber - prevBlockNumber

	// we stop when the range we are in-between is 1 or 0 (same block), it means we found a block with the exact timestamp or lowest from below
	if blockNumberDiff == 1 || blockNumberDiff == 0 {
		// if the block we are returning has a ts > target, it means we want one block before (so our ts is always bigger than block ts)
		if currentTimestampSeconds > timestampSeconds {
			return big.NewInt(currentBlockNumber - 1), nil
		} else {
			return big.NewInt(currentBlockNumber), nil
		}
	}

	timeDiff := currentTimestampSeconds - prevTimestampSeconds
	secondsPerBlock := int64(math.Ceil(float64(timeDiff) / float64(blockNumberDiff)))
	distanceToTargetFromCurrent := currentTimestampSeconds - timestampSeconds
	blocksToJump := distanceToTargetFromCurrent / secondsPerBlock
	f.logger.Info("eth block search delta", log.Int64("jump-backwards", blocksToJump))
	guessBlockNumber := currentBlockNumber - blocksToJump
	guess, err := f.btg.ApproximateBlockAt(ctx, big.NewInt(guessBlockNumber))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to get header by block number %d", guessBlockNumber))
	}

	return f.findBlockByTimeStamp(ctx, timestampSeconds, guess.Number, guess.TimeSeconds, currentBlockNumber, currentTimestampSeconds)
}
