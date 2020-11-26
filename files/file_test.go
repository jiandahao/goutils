package files

import (
	"testing"
)

func TestReadLines(t *testing.T) {
	lines, err := ReadLines("./files.go")
	if err != nil {
		t.Error(err)
		return
	}

	for {
		line, ok := <-lines
		if !ok {
			t.Log("done")
			return
		}
		_ = line
	}
}
