package precheck

var Latest = "v3.2.0"

var matrix = map[string]map[string]map[string]bool{
	"v3.2.0": {
		"k8s": {
			"v1.22": true,
			"v1.21": true,
			"v1.20": true,
			"v1.19": true,
		},
		"ks": {
			"v3.1.1": true,
			"v3.1.0": true,
		},
	},
	"v3.1.1": {
		"k8s": {
			"v1.20": true,
			"v1.19": true,
			"v1.18": true,
			"v1.17": true,
			"v1.16": true,
			"v1.15": true,
		},
		"ks": {
			"v3.0.0": true,
			"v3.1.0": true,
		},
	},
	"v3.1.0": {
		"k8s": {
			"v1.20": true,
			"v1.19": true,
			"v1.18": true,
			"v1.17": true,
			"v1.16": true,
			"v1.15": true,
		},
		"ks": {
			"v3.0.0": true,
		},
	},
	"v3.0.0": {
		"k8s": {
			"v1.18": true,
			"v1.17": true,
			"v1.16": true,
			"v1.15": true,
		},
		"ks": {
			"v2.1.1": true,
		},
	},
	"v2.1.1": {
		"k8s": {
			"v1.18": true,
			"v1.17": true,
			"v1.16": true,
			"v1.15": true,
		},
		"ks": {
			"v2.1.1": true,
		},
	},
	"other": {
		"k8s": {
			"v1.22": true,
			"v1.21": true,
			"v1.20": true,
			"v1.19": true,
			"v1.18": true,
			"v1.17": true,
			"v1.16": true,
			"v1.15": true,
		},
	},
}
