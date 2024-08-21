package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"

	kkprojectv1 "github.com/kubesphere/kubekey/v4/pkg/apis/project/v1"
)

func TestBlockExecutor_DealRunOnce(t *testing.T) {
	testcases := []struct {
		name    string
		runOnce bool
		except  []string
	}{
		{
			name:    "runonce is false",
			runOnce: false,
			except:  []string{"node1", "node2", "node3"},
		},
		{
			name:    "runonce is true",
			runOnce: true,
			except:  []string{"node1"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, blockExecutor{
				hosts: []string{"node1", "node2", "node3"},
			}.dealRunOnce(tc.runOnce), tc.except)
		})
	}
}

func TestBlockExecutor_DealIgnoreErrors(t *testing.T) {
	testcases := []struct {
		name         string
		ignoreErrors *bool
		except       *bool
	}{
		{
			name:         "ignoreErrors is empty",
			ignoreErrors: nil,
			except:       ptr.To(true),
		},
		{
			name:         "ignoreErrors is true",
			ignoreErrors: ptr.To(true),
			except:       ptr.To(true),
		},
		{
			name:         "ignoreErrors is false",
			ignoreErrors: ptr.To(false),
			except:       ptr.To(false),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, blockExecutor{
				ignoreErrors: ptr.To(true),
			}.dealIgnoreErrors(tc.ignoreErrors), tc.except)
		})
	}
}

func TestBlockExecutor_DealTags(t *testing.T) {
	testcases := []struct {
		name   string
		tags   kkprojectv1.Taggable
		except kkprojectv1.Taggable
	}{
		{
			name:   "single tags",
			tags:   kkprojectv1.Taggable{Tags: []string{"c"}},
			except: kkprojectv1.Taggable{Tags: []string{"a", "b", "c"}},
		},
		{
			name:   "mutil tags",
			tags:   kkprojectv1.Taggable{Tags: []string{"c", "d"}},
			except: kkprojectv1.Taggable{Tags: []string{"a", "b", "c", "d"}},
		},
		{
			name:   "repeat tags",
			tags:   kkprojectv1.Taggable{Tags: []string{"b", "c"}},
			except: kkprojectv1.Taggable{Tags: []string{"a", "b", "c"}},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, blockExecutor{
				tags: kkprojectv1.Taggable{Tags: []string{"a", "b"}},
			}.dealTags(tc.tags).Tags, tc.except.Tags)
		})
	}
}

func TestBlockExecutor_DealWhen(t *testing.T) {
	testcases := []struct {
		name   string
		when   []string
		except []string
	}{
		{
			name:   "single when",
			when:   []string{"c"},
			except: []string{"a", "b", "c"},
		},
		{
			name:   "mutil when",
			when:   []string{"c", "d"},
			except: []string{"a", "b", "c", "d"},
		},
		{
			name:   "repeat when",
			when:   []string{"b", "c"},
			except: []string{"a", "b", "c"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, blockExecutor{
				when: []string{"a", "b"},
			}.dealWhen(kkprojectv1.When{Data: tc.when}), tc.except)
		})
	}
}
