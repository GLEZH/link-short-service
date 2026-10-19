package pool

import "testing"

type item struct {
	name       string
	values     []int
	attributes map[string]string
	resetCalls int
}

func (i *item) Reset() {
	i.name = ""
	i.values = i.values[:0]
	clear(i.attributes)
	i.resetCalls++
}

func TestPool(t *testing.T) {
	created := 0
	p := New(func() *item {
		created++
		return &item{
			values:     make([]int, 0, 4),
			attributes: make(map[string]string),
		}
	})

	value := p.Get()
	value.name = "value"
	value.values = append(value.values, 1, 2)
	value.attributes["key"] = "value"
	valuesCapacity := cap(value.values)
	attributes := value.attributes

	p.Put(value)

	if value.name != "" {
		t.Errorf("name = %q, want empty string", value.name)
	}
	if len(value.values) != 0 {
		t.Errorf("values length = %d, want 0", len(value.values))
	}
	if cap(value.values) != valuesCapacity {
		t.Errorf("values capacity = %d, want %d", cap(value.values), valuesCapacity)
	}
	if len(value.attributes) != 0 {
		t.Errorf("attributes length = %d, want 0", len(value.attributes))
	}
	if value.attributes == nil {
		t.Error("attributes map is nil")
	}
	attributes["reused"] = "true"
	if value.attributes["reused"] != "true" {
		t.Error("attributes map was replaced")
	}
	if value.resetCalls != 1 {
		t.Errorf("Reset calls = %d, want 1", value.resetCalls)
	}

	if got := p.Get(); got == nil {
		t.Fatal("Get returned nil")
	}
	if created < 1 {
		t.Fatal("constructor was not called")
	}
}

func TestNewPanicsForNilConstructor(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("New did not panic")
		}
	}()

	New[*item](nil)
}
