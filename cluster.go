package gocbt

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gocb "gopkg.in/couchbase/gocb.v1"
)

// Cluster provides access to cluster-level configuration and operations.
type Cluster struct {
	ip       string
	username string
	password string
}

func (c *Cluster) connect(t *testing.T) *gocb.Cluster {
	cli, err := gocb.Connect("couchbase://" + c.ip)
	require.NoError(t, err)

	err = cli.Authenticate(gocb.PasswordAuthenticator{
		Username: c.username,
		Password: c.password,
	})
	require.NoError(t, err)

	return cli
}

func (c *Cluster) configure(t *testing.T, opts ...ClusterConfigOption) {
	conf := defaultClusterConfig()
	for _, opt := range opts {
		opt(&conf)
	}

	c.createBuckets(t, conf.buckets, conf.bucketTimeout)
	c.createIndexes(t, conf.indexes, conf.indexTimeout)
}

func (c *Cluster) createBuckets(t *testing.T, buckets map[string]bucketConfig, timeout int) {
	cli := c.connect(t)
	defer func() {
		err := cli.Close()
		require.NoError(t, err)
	}()

	for name, conf := range buckets {
		err := cli.Manager(c.username, c.password).InsertBucket(&gocb.BucketSettings{
			Name:  name,
			Quota: conf.quotaMb,
		})
		require.NoError(t, err)

		secs := 0
		for {
			if _, err := cli.OpenBucket(name, ""); err == nil {
				return
			}

			if timeout > 0 && secs >= timeout {
				t.Fatalf("timed out waiting for bucket %s for %d seconds; try increasing the timeout using BucketTimeout(#)", name, timeout)
			} else {
				time.Sleep(1 * time.Second)
				secs++
			}
		}
	}
}

func (c *Cluster) createIndexes(t *testing.T, indexes []indexConfig, timeout int) {
	cli := c.connect(t)
	defer func() {
		err := cli.Close()
		require.NoError(t, err)
	}()

	for _, index := range indexes {
		secs := 0
		for {
			if _, err := cli.OpenBucket(index.bucket, ""); err != nil {
				if strings.ToLower(err.Error()) == "no access" {
					goto sleep
				} else {
					require.NoError(t, err)
				}
			}

			if err := cli.Manager(c.username, c.password).SearchIndexManager().CreateIndex(index.builder); err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "no available fts nodes") {
					goto sleep
				} else {
					require.NoError(t, err)
				}
			}

			return

		sleep:
			if timeout > 0 && secs >= timeout {
				t.Fatalf("timed out waiting for index %s in bucket %s for %d seconds; try increasing the timeout using IndexTimeout(#)", index.name, index.bucket, timeout)
			} else {
				time.Sleep(1 * time.Second)
				secs++
			}
		}
	}
}

// ClusterConfigOption functions can be passed to Node.Configure() to configure its creation.
type ClusterConfigOption func(*clusterConfig)

// Bucket configures a bucket creation with the specified name, and other optional parameters.
func Bucket(name string, opts ...BucketConfigOption) ClusterConfigOption {
	return func(conf *clusterConfig) {
		bConf := defaultBucketConfig()
		for _, opt := range opts {
			opt(&bConf)
		}
		conf.buckets[name] = bConf
	}
}

// BucketTimeout configures a time in seconds to wait for a bucket to be created.
func BucketTimeout(secs int) ClusterConfigOption {
	return func(conf *clusterConfig) {
		conf.bucketTimeout = secs
	}
}

// GeoSearchIndex configures a geospatial index creation with the specified name for the specified bucket.
func GeoSearchIndex(name, bucket string) ClusterConfigOption {
	return func(conf *clusterConfig) {
		b := gocb.SearchIndexDefinitionBuilder{}
		b.AddField("name", name)
		b.AddField("sourceName", bucket)
		b.AddField("params", map[string]interface{}{
			"doc_config": map[string]interface{}{
				"docid_prefix_delim": "",
				"docid_regexp":       "",
				"mode":               "type_field",
				"type_field":         "type",
			},
			"mapping": map[string]interface{}{
				"index_dynamic": true,
				"default_mapping": map[string]interface{}{
					"properties": map[string]interface{}{
						"geo": map[string]interface{}{
							"fields": []interface{}{
								map[string]interface{}{
									"index":                true,
									"name":                 "geo",
									"store":                true,
									"type":                 "geopoint",
									"include_in_all":       true,
									"include_term_vectors": true,
								},
							},
							"dynamic": false,
							"enabled": true,
						},
					},
					"dynamic": true,
					"enabled": true,
				},
				"default_analyzer":        "standard",
				"default_datetime_parser": "dateTimeOptional",
				"default_field":           "_all",
				"default_type":            "_default",
				"docvalues_dynamic":       true,
				"store_dynamic":           false,
				"type_field":              "_type",
				"analysis":                map[string]interface{}{},
			},
			"store": map[string]interface{}{
				"indexType":   "scorch",
				"kvStoreName": "",
			},
		})
		b.AddField("type", "fulltext-index")
		b.AddField("sourceType", "couchbase")

		conf.indexes = append(conf.indexes, indexConfig{
			builder: b,
			name:    name,
			bucket:  bucket,
		})
	}
}

// IndexTimeout configures a time in seconds to wait for an index to be created.
func IndexTimeout(secs int) ClusterConfigOption {
	return func(conf *clusterConfig) {
		conf.indexTimeout = secs
	}
}

type clusterConfig struct {
	buckets       map[string]bucketConfig
	bucketTimeout int

	indexes      []indexConfig
	indexTimeout int
}

type indexConfig struct {
	builder gocb.SearchIndexDefinitionBuilder
	name    string
	bucket  string
}

func defaultClusterConfig() clusterConfig {
	return clusterConfig{
		buckets:       make(map[string]bucketConfig),
		bucketTimeout: 10,
		indexTimeout:  20,
	}
}
