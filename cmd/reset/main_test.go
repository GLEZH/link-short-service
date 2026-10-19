package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "go.mod"), "module example.com/resettest\n\ngo 1.26\n")
	writeTestFile(t, filepath.Join(root, "sample", "sample.go"), `package sample

import "time"

type Child struct {
	Count int
}

func (v *Child) Reset() {
	v.Count = 0
}

func resetGenZero() {}

// generate:reset
type GeneratedChild struct {
	Name string
}

// generate:reset
type Sample[T ~int] struct {
	Integer        int
	Text           string
	Enabled        bool
	Number         *int
	Numbers        []int
	NumbersPointer *[]int
	Lookup         map[string]int
	LookupPointer  *map[string]int
	Child          Child
	ChildPointer   *Child
	Generated      GeneratedChild
	GeneratedPtr   *GeneratedChild
	Stamp          time.Time
	Generic        T
	PointerPointer **int
	Nothing        any
}

type Unmarked struct {
	Value int
}
`)
	writeTestFile(t, filepath.Join(root, "sample", "sample_test.go"), `package sample

import (
	"testing"
	"time"
)

func TestSampleReset(t *testing.T) {
	number := 42
	pointerNumber := 24
	pointer := &pointerNumber
	numbers := make([]int, 2, 8)
	pointerNumbers := make([]int, 2, 6)
	pointerLookup := map[string]int{"key": 2}
	value := Sample[int]{
		Integer:        1,
		Text:           "value",
		Enabled:        true,
		Number:         &number,
		Numbers:        numbers,
		NumbersPointer: &pointerNumbers,
		Lookup:         map[string]int{"key": 1},
		LookupPointer:  &pointerLookup,
		Child:          Child{Count: 2},
		ChildPointer:   &Child{Count: 3},
		Generated:      GeneratedChild{Name: "child"},
		GeneratedPtr:   &GeneratedChild{Name: "child pointer"},
		Stamp:          time.Now(),
		Generic:        4,
		PointerPointer: &pointer,
		Nothing:        "value",
	}

	numberAddress := value.Number
	childAddress := value.ChildPointer
	generatedAddress := value.GeneratedPtr
	pointerAddress := value.PointerPointer
	value.Reset()

	if value.Integer != 0 || value.Text != "" || value.Enabled {
		t.Error("primitive values were not reset")
	}
	if value.Number != numberAddress || *value.Number != 0 {
		t.Error("pointer value was not reset")
	}
	if value.Numbers == nil || len(value.Numbers) != 0 || cap(value.Numbers) != 8 {
		t.Error("slice was not truncated")
	}
	if value.NumbersPointer == nil || len(*value.NumbersPointer) != 0 || cap(*value.NumbersPointer) != 6 {
		t.Error("slice pointer was not truncated")
	}
	if value.Lookup == nil || len(value.Lookup) != 0 {
		t.Error("map was not cleared")
	}
	if value.LookupPointer == nil || *value.LookupPointer == nil || len(*value.LookupPointer) != 0 {
		t.Error("map pointer was not cleared")
	}
	if value.Child.Count != 0 || value.ChildPointer != childAddress || value.ChildPointer.Count != 0 {
		t.Error("nested structures were not reset")
	}
	if value.Generated.Name != "" || value.GeneratedPtr != generatedAddress || value.GeneratedPtr.Name != "" {
		t.Error("generated nested structures were not reset")
	}
	if !value.Stamp.IsZero() || value.Generic != 0 || value.Nothing != nil {
		t.Error("zero values were not assigned")
	}
	if value.PointerPointer != pointerAddress || **value.PointerPointer != 0 {
		t.Error("nested pointer value was not reset")
	}
}
`)

	if err := generate(root); err != nil {
		t.Fatalf("generate() error = %v", err)
	}

	generatedPath := filepath.Join(root, "sample", generatedFileName)
	first, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	for _, expected := range []string{
		"func (v *Sample[T]) Reset()",
		"func (v *GeneratedChild) Reset()",
		"v.Integer = 0",
		`v.Text = ""`,
		"v.Numbers = v.Numbers[:0]",
		"clear(v.Lookup)",
		"v.Child.Reset()",
		"(*v.ChildPointer).Reset()",
		"v.Generated.Reset()",
		"(*v.GeneratedPtr).Reset()",
		"resetGenZero2(&v.Stamp)",
	} {
		if !strings.Contains(string(first), expected) {
			t.Errorf("generated code does not contain %q", expected)
		}
	}
	if strings.Contains(string(first), "Unmarked") {
		t.Error("generated code contains an unmarked structure")
	}

	if err = generate(root); err != nil {
		t.Fatalf("second generate() error = %v", err)
	}
	second, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(first) != string(second) {
		t.Error("generated code is not stable")
	}

	command := exec.Command("go", "test", "./...")
	command.Dir = root
	command.Env = append(os.Environ(), "GOWORK=off")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generated code test error = %v\n%s", err, output)
	}
}

func TestModuleRoot(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n")
	dir := filepath.Join(root, "nested", "package")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	t.Chdir(dir)

	got, err := moduleRoot()
	if err != nil {
		t.Fatalf("moduleRoot() error = %v", err)
	}
	if got != root {
		t.Errorf("moduleRoot() = %q, want %q", got, root)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
