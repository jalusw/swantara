package helper

import "testing"

func TestAssertStatusPasses(t *testing.T) {
	AssertStatus(t, 200, 200)
}
