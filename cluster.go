package gocbt

import (
	"testing"

	"github.com/stretchr/testify/require"
	gocb "gopkg.in/couchbase/gocb.v1"
)

type Cluster struct {
	cli      *gocb.Cluster
	username string
	password string
}

func connectToCluster(t *testing.T, ip, username, password string) *Cluster {
	cli, err := gocb.Connect("couchbase://" + ip)
	require.NoError(t, err)
	return &Cluster{
		cli:      cli,
		username: username,
		password: password,
	}
}

func (c *Cluster) configure(t *testing.T, opts ...ClusterConfigOption) {
	conf := defaultClusterConfig()
	for _, opt := range opts {
		opt(&conf)
	}

	for name, conf := range conf.buckets {
		c.createBucket(t, name, conf)
	}
}

func (c *Cluster) createBucket(t *testing.T, name string, conf bucketConfig) {
	err := c.cli.Manager(c.username, c.password).InsertBucket(&gocb.BucketSettings{
		Name:  name,
		Quota: conf.quotaMb,
	})
	require.NoError(t, err)
}

func (c *Cluster) close(t *testing.T) {
	err := c.cli.Close()
	require.NoError(t, err)
}

type ClusterConfigOption func(*clusterConfig)

func Bucket(name string, opts ...BucketConfigOption) ClusterConfigOption {
	return func(conf *clusterConfig) {
		bConf := defaultBucketConfig()
		for _, opt := range opts {
			opt(&bConf)
		}
		conf.buckets[name] = bConf
	}
}

type clusterConfig struct {
	buckets map[string]bucketConfig
}

func defaultClusterConfig() clusterConfig {
	return clusterConfig{
		buckets: make(map[string]bucketConfig),
	}
}
