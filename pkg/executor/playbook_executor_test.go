package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaybookExecutor_DealRunOnce(t *testing.T) {
	testcases := []struct {
		name       string
		runOnce    bool
		hosts      []string
		batchHosts [][]string
		except     [][]string
	}{
		{
			name:       "runonce is false",
			runOnce:    false,
			batchHosts: [][]string{{"node1", "node2"}},
			except:     [][]string{{"node1", "node2"}},
		},
		{
			name:       "runonce is true",
			runOnce:    true,
			hosts:      []string{"node1"},
			batchHosts: [][]string{{"node1", "node2"}},
			except:     [][]string{{"node1"}},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			playbookExecutor{}.dealRunOnce(tc.runOnce, tc.hosts, &tc.batchHosts)
			assert.Equal(t, tc.batchHosts, tc.except)
		})
	}
}
