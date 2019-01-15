package gocbt

type BucketConfigOption func(*bucketConfig)

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
