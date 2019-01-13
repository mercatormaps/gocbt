package gocbt_test

import (
	"testing"

	"github.com/joe-mann/gocbt"
)

func TestExampleNode(t *testing.T) {
	n := gocbt.NewNode()
	defer n.Teardown(t)
	n.Setup(t)
}
