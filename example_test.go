package gocbt_test

import (
	"testing"

	"github.com/mercatormaps/gocbt"
)

func TestExampleNode(t *testing.T) {
	gocbt.EnableLogging()
	n := gocbt.NewNode()
	defer n.Teardown(t)
	n.Setup(t).Configure(t,
		gocbt.Bucket("sample_bucket_1"),
		gocbt.Bucket("sample_bucket_2"),
		gocbt.GeoSearchIndex("my-geo", "sample_bucket_1"))
}
