package consensus

import (
	"context"
	"github.com/pkg/errors"
	"reflect"
)

type Consensus struct {
	parameters Parameters
	preference []byte
	confidence int
	isRunning  bool
}

func NewConsensus(parameters Parameters, preference []byte) (*Consensus, error) {
	err := parameters.Verify()
	if err != nil {
		return nil, errors.Wrap(err, "unable to verify the consensus configuration")
	}

	consensus := &Consensus{
		parameters: parameters,
		confidence: 0,
		isRunning:  false,
		preference: preference,
	}

	return consensus, nil
}

// Sync synchronize data between the peers
//
// Ref: https://github.com/ava-labs/mastering-avalanche/blob/main/chapter_09.md
func (c *Consensus) Sync(ctx context.Context, setNewBlockDataFunc func([]byte) error, getBlockDataFromRandomKFunc func(int) ([][]byte, error)) error {
	if c.isRunning {
		return errors.New("consensus is running")
	}
	c.isRunning = true
	c.confidence = 1
	for c.confidence < c.parameters.Beta {
		if !c.isRunning {
			break
		}
		// ask k random peers to get the preferences
		preferenceFromK, err := getBlockDataFromRandomKFunc(c.parameters.K)
		if err != nil {
			return errors.Wrap(err, "unable to get get block data from cb function")
		}
		if len(preferenceFromK) < c.parameters.K {
			continue
		}
		frequent, preference, err := c.GetMostFrequentPreference(preferenceFromK)
		if err != nil {
			return errors.Wrap(err, "unable to get the most frequent")
		}
		// if the most frequent item is larger α
		if frequent >= c.parameters.Alpha {
			oldPreference := c.preference
			c.preference = preference
			// set the current data block to the new preference
			err := setNewBlockDataFunc(c.preference)
			if err != nil {
				return errors.Wrap(err, "unable to update the preference")
			}

			if reflect.DeepEqual(oldPreference, c.preference) {
				c.confidence++
			} else {
				c.confidence = 1
			}
		} else {
			c.confidence = 0
		}
	}
	c.isRunning = false
	return nil
}

func (c *Consensus) GetMostFrequentPreference(preferences [][]byte) (int, []byte, error) {
	if len(preferences) == 0 {
		return 0, nil, errors.New("the preferences is empty")
	}
	var count int
	var preference []byte
	for _, preference1 := range preferences {
		maxCount := 0
		for _, preference2 := range preferences {
			if reflect.DeepEqual(preference1, preference2) {
				maxCount++
			}
		}
		if maxCount > count {
			count = maxCount
			preference = preference1
		}
	}
	return count, preference, nil
}
