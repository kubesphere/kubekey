package render

import (
	"fmt"
	"testing"
	"text/template"
)

func TestRender(t *testing.T) {
	type args struct {
		input       string
		replaceVars map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				input: "kubeadm join {{ ApiServer }} --token {{ Token }} --discovery-token-ca-cert-hash {{ Hash }}",
				replaceVars: map[string]string{
					"ApiServer": "192.168.0.1",
					"Token":     "tokentokentoken",
					"Hash":      "hashhashhashhashhashhash",
				},
			},
			want: "kubeadm join 192.168.0.1 --token tokentokentoken --discovery-token-ca-cert-hash hashhashhashhashhashhash",
		},
		{
			name: "test2",
			args: args{
				input:       "kubeadm init -f kubeadm.conf",
				replaceVars: nil,
			},
			want: "kubeadm init -f kubeadm.conf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.args.input, tt.args.replaceVars)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubeKeyTemplate_PauseRender(t *testing.T) {
	addNodeCmdTmpl := template.Must(template.New("").Parse(
		"kubeadm join {{ .ApiServer }} --token {{ .Token }} --discovery-token-ca-cert-hash {{ .Hash }}",
	))

	kkTmpl := NewKebeKeyTemplate(addNodeCmdTmpl)
	v1 := map[string]interface{}{
		"ApiServer": "127.0.0.1",
	}

	v2 := map[string]interface{}{
		"Token": "tokentokentoken",
	}

	v3 := map[string]interface{}{
		"Hash": "hashhashhash",
	}

	//kkTmpl = kkTmpl.PauseRender(v1)
	//fmt.Println(kkTmpl.String())

	kkTmpl.PauseRender(v1)

	fmt.Println(kkTmpl.String())

	kkTmpl.PauseRender(v2)
	kkTmpl.PauseRender(v3)
	fmt.Println(kkTmpl.String())
}
