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
				"v1.15.12": "e052bae41e731921a9197b4d078c30d33fac5861716dc275bfee4670addbac9b",
				"v1.16.10": "726d42c569f25078d03b758477f17f543c845aef2ff48acd9d4269705ca1aa9d",
				"v1.17.4":  "3cdcffcf8a1660241a045cfdfed3ebbf7f7c6a0840f008e2b049b533bca5bb8c",
				"v1.17.5":  "9bd2fd1118b3d07d12e2a806c04bf34d99e79886c5318ddc003ba38f30da390c",
				"v1.17.6":  "d4cfc9a0a734ba015594974ee4253b8965b95cdb6e83d8a6a946675aad418b40",
				"v1.18.3":  "a60974e9840e006076d204fd4ddcba96213beba10fb89ff01882095546c9684d",
			},
		},
		kubelet: {
			amd64: {
				"v1.15.12": "dff48393a3116b8f7dea206b81678e52f7fad298f1aff976f18f1bfa4e9ccdde",
				"v1.16.10": "82b38f444d11c2436040165b1addf46d0909a6daec9133cc979678835ef8e14b",
				"v1.17.4":  "f3a427ddf610b568db60c8d47565041901220e1bbe257614b61bb4c76801d765",
				"v1.17.5":  "c5fbfa83444bdeefb51934c29f0b4b7ffc43ce5a98d7f957d8a11e3440055383",
				"v1.17.6":  "4b7fd5123bfafe2249bf91ed83469c2655a8d3295966e5fbd952f89b64b75f57",
				"v1.18.3":  "6aac8853028a4f185de5ccb5b41b3fbd87726161445dee56f351e3e51442d669",
			},
		},
		kubectl: {
			amd64: {
				"v1.15.12": "a32b762279c33cb8d8f4198f3facdae402248c3164e9b9b664c3afbd5a27472e",
				"v1.16.10": "246d36e4ce67e74e95ff2ba578b9189f58e5def0e8830a24cd30fa3cf279742f",
				"v1.17.4":  "465b2d2bd7512b173860c6907d8127ee76a19a385aa7865608e57a5eebe23597",
				"v1.17.5":  "03cd1fa19f90d38005148793efdb17a9b58d01dedea641a8496b9cf228db3ab4",
				"v1.17.6":  "5e245f6af6fb761fbe4b3ac06b753f33b361ce0486c48c85b45731a7ee5e4cca",
				"v1.18.3":  "6fcf70aae5bc64870c358fac153cdfdc93f55d8bae010741ecce06bb14c083ea",
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
