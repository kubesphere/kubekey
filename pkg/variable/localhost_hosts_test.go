package variable_test

import (
	"context"
	"testing"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

func TestNewResolvesLocalhostHost(t *testing.T) {
	client, playbook, err := _const.NewTestPlaybook([]string{"test"})
	if err != nil {
		t.Fatal(err)
	}
	v, err := variable.New(context.Background(), client, *playbook, source.MemorySource)
	if err != nil {
		t.Fatal(err)
	}
	got, err := v.Get(variable.GetHostnames([]string{_const.VariableLocalHost}))
	if err != nil {
		t.Fatal(err)
	}
	hs, ok := got.([]string)
	if !ok || len(hs) != 1 || hs[0] != _const.VariableLocalHost {
		t.Fatalf("got %v, want [localhost]", got)
	}
}
