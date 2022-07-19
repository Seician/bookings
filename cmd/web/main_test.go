package main

import "testing"

func Test_run(t *testing.T) {

	err := run()
	if err != nil {
		t.Errorf("failed run")
	}
}
