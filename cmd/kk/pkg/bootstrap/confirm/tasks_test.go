package confirm

import "testing"

func TestRefineDockerVersion(t *testing.T) {
	cases := map[string]string{
		"1.2.3":    "1.2.3",
		"23.0.1":   "23.0.1",
		" 1.2.3 ":  "1.2.3",
		" 23.0.1 ": "23.0.1",
		"abc":      "err",
	}

	for k, v := range cases {
		result, err := RefineDockerVersion(k)
		if err != nil {
			if v != "err" {
				t.Errorf("case \"%s\", result \"%s\" expected, but \"%s\" get", k, v, "err")
			}
		} else {
			if result != v {
				t.Errorf("case \"%s\", result \"%s\" expected, but \"%s\" get", k, v, result)
			}
		}
	}
}
