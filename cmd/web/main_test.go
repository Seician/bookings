package main

import "testing"

func Test_run(t *testing.T) {

	_, err := run()
	if err != nil {
		t.Errorf("failed run")
	}
}
