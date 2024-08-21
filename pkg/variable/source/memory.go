package source

var _ Source = &memorySource{}

type memorySource struct {
	data map[string][]byte
}

// NewMemorySource returns a new memorySource.
func NewMemorySource() Source {
	return &memorySource{
		data: make(map[string][]byte),
	}
}

func (m *memorySource) Read() (map[string][]byte, error) {
	return m.data, nil
}

func (m *memorySource) Write(data []byte, filename string) error {
	m.data[filename] = data

	return nil
}
