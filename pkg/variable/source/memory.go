package source

var _ Source = &memorySource{}

type memorySource struct {
	data map[string]map[string]any
}

// NewMemorySource returns a new memorySource.
func NewMemorySource() Source {
	return &memorySource{
		data: make(map[string]map[string]any),
	}
}

func (m *memorySource) Read() (map[string]map[string]any, error) {
	return m.data, nil
}

func (m *memorySource) Write(data map[string]any, filename string) error {
	m.data[filename] = data

	return nil
}
