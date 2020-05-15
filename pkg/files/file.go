package files

const (
	kubeadm = "kubeadm"
	kubelet = "kubelet"
	kubectl = "kubectl"
	kubecni = "kubecni"
	helm    = "helm"
	amd64   = "amd64"
	arm64   = "arm64"
)

type KubeBinary struct {
	Name    string
	Arch    string
	Version string
	Url     string
	Path    string
	GetCmd  string
}

var (
	fileSha256 = map[string]map[string]map[string]string{
		kubeadm: {
			amd64: {
				"v1.17.4": "3cdcffcf8a1660241a045cfdfed3ebbf7f7c6a0840f008e2b049b533bca5bb8c",
				"v1.17.5": "9bd2fd1118b3d07d12e2a806c04bf34d99e79886c5318ddc003ba38f30da390c",
			},
		},
		kubelet: {
			amd64: {
				"v1.17.4": "f3a427ddf610b568db60c8d47565041901220e1bbe257614b61bb4c76801d765",
				"v1.17.5": "c5fbfa83444bdeefb51934c29f0b4b7ffc43ce5a98d7f957d8a11e3440055383",
			},
		},
		kubectl: {
			amd64: {
				"v1.17.4": "465b2d2bd7512b173860c6907d8127ee76a19a385aa7865608e57a5eebe23597",
				"v1.17.5": "03cd1fa19f90d38005148793efdb17a9b58d01dedea641a8496b9cf228db3ab4",
			},
		},
		helm: {
			amd64: {
				"v3.2.1": "98c57f2b86493dd36ebaab98990e6d5117510f5efbf21c3344c3bdc91a4f947c",
			},
		},
		kubecni: {
			amd64: {
				"v0.8.2": "21283754ffb953329388b5a3c52cef7d656d535292bda2d86fcdda604b482f85",
				"v0.8.6": "994fbfcdbb2eedcfa87e48d8edb9bb365f4e2747a7e47658482556c12fd9b2f5",
			},
		},
	}
)

func (binary *KubeBinary) GetSha256() string {
	sha256 := fileSha256[binary.Name][binary.Arch][binary.Version]
	return sha256
}
