package gocbt

// BucketConfigOption functions can be passed to Bucket() to configure its creation.
type BucketConfigOption func(*bucketConfig)

// MemoryQuota for a bucket in megabytes.
func MemoryQuota(mb int) BucketConfigOption {
	return func(conf *bucketConfig) {
		conf.quotaMb = mb
	}
}

type bucketConfig struct {
	quotaMb int
}

func defaultBucketConfig() bucketConfig {
	return bucketConfig{
		quotaMb: 128,
	}
}
