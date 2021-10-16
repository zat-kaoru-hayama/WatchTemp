package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSystem(t *testing.T) {
	workpath := filepath.Join(os.TempDir(), "work.txt")

	process, err := System(fmt.Sprintf(`echo "ahaha" > "%s"`, workpath))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	process.Wait()

	output, err := os.ReadFile(workpath)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if string(output) != "\"ahaha\" \r\n" {
		t.Fatalf("result = `%s` (expect \"ahaha\")", string(output))
		return
	}
	os.Remove(workpath)
}
