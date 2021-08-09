package main

type MockBackend struct {
	store map[string]string
}

func NewMockBackend() MockBackend {
	var backend MockBackend
	backend.store = make(map[string]string)
	return backend
}

func (e *MockBackend) Get(key string) (string, bool, error) {
	etcdKey := etcdItemPrefix + key

	v, ok := e.store[etcdKey]
	return v, ok, nil
}

func (e *MockBackend) Set(key string, value string) error {
	etcdKey := etcdItemPrefix + key

	e.store[etcdKey] = value

	return nil
}

func (e *MockBackend) Delete(key string) error {
	etcdKey := etcdItemPrefix + key

	delete(e.store, etcdKey)

	return nil
}
