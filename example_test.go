package gocbt_test

import (
	"testing"
	"time"

	"github.com/couchbase/gocb/cbft"
	"github.com/mercatormaps/gocbt"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	gocb "gopkg.in/couchbase/gocb.v1"
)

type ExampleTestSuite struct {
	suite.Suite
	couchbase *gocbt.Node
}

func (s *ExampleTestSuite) SetupSuite() {
	gocbt.EnableLogging()

	s.couchbase = gocbt.NewNode()
	s.couchbase.Setup(s.T()).Configure(s.T(),
		gocbt.Bucket("sample_bucket_1"),
		gocbt.Bucket("sample_bucket_2"),
		gocbt.GeoSearchIndex("my-geo", "geometry.coordinates", "sample_bucket_1"))
}

func (s *ExampleTestSuite) TearDownSuite() {
	if s.couchbase != nil {
		s.couchbase.Teardown(s.T())
	}
}

func (s *ExampleTestSuite) TestExampleNode() {
	cluster, err := gocb.Connect(s.couchbase.Host())
	require.NoError(s.T(), err)

	err = cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: s.couchbase.Username(),
		Password: s.couchbase.Password(),
	})
	require.NoError(s.T(), err)

	bucket, err := cluster.OpenBucket("sample_bucket_1", "")
	require.NoError(s.T(), err)

	type Point struct {
		Geometry struct {
			Coordinates []int `json:"coordinates"`
		} `json:"geometry"`
	}

	_, err = bucket.Insert("my-point", struct {
		Geometry map[string]interface{} `json:"geometry"`
	}{
		Geometry: map[string]interface{}{
			"type":        "Point",
			"coordinates": []int{100, 30},
		},
	}, 0)
	require.NoError(s.T(), err)

	point := Point{}
	_, err = bucket.Get("my-point", &point)
	require.NoError(s.T(), err)

	time.Sleep(10 * time.Second)

	query := gocb.NewSearchQuery("my-geo", cbft.NewGeoBoundingBoxQuery(90, -180, -90, 180))
	results, err := bucket.ExecuteSearchQuery(query)
	require.NoError(s.T(), err)
	require.Len(s.T(), results.Hits(), 1)
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(ExampleTestSuite))
}
