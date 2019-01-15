package gocbt_test

import (
	"testing"

	"github.com/joe-mann/gocbt"
)

func TestExampleNode(t *testing.T) {
	n := gocbt.NewNode()
	defer n.Teardown(t)
	n.Setup(t)
	n.Configure(t,
		gocbt.Bucket("sample_bucket_1"),
		gocbt.Bucket("sample_bucket_2"))
}
