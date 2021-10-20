package v1alpha2

type Event struct {
	Step    string `yaml:"step" json:"step,omitempty"`
	Status  string `yaml:"status" json:"status,omitempty"`
	Message string `yaml:"message" json:"message,omitempty"`
}
