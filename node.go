package gocbt

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Node struct {
	containerID string
	host        string
	username    string
	password    string
}

func NewNode() *Node {
	return &Node{}
}

func (n *Node) Setup(t *testing.T, opts ...NodeConfigOption) {
	conf := defaultNodeConfig()
	for _, opt := range opts {
		opt(&conf)
	}

	id, ip := pullAndStart(t, conf.image)
	n.containerID = id

	wait(t, ip, conf.timeoutSecs)

	setMemoryQuotas(t, ip,
		strconv.FormatUint(uint64(conf.dataQuotaMb), 10),
		strconv.FormatUint(uint64(conf.indexQuotaMb), 10),
		strconv.FormatUint(uint64(conf.searchQuotaMb), 10))

	port := strconv.FormatUint(uint64(conf.port), 10)
	setPortAndCredentials(t, ip, port, conf.username, conf.password)
	n.host = "couchbase://" + ip + ":" + port
	n.username = conf.username
	n.password = conf.password
}

func (n *Node) Teardown(t *testing.T) {
	if n.containerID != "" {
		stopAndRemove(t, n.containerID)
	}
}

type NodeConfigOption func(*nodeConfig)

func DockerImage(image string) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.image = image
	}
}

func Timeout(secs int) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.timeoutSecs = secs
	}
}

func Port(port uint) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.port = port
	}
}

func MemoryQuotas(dataMb, indexMb, searchMb uint) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.dataQuotaMb = dataMb
		conf.indexQuotaMb = indexMb
		conf.searchQuotaMb = searchMb
	}
}

func Credentials(username, password string) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.username = username
		conf.password = password
	}
}

func Bucket(name string) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.buckets = append(conf.buckets, name)
	}
}

type nodeConfig struct {
	image       string
	timeoutSecs int

	port     uint
	username string
	password string

	dataQuotaMb   uint
	indexQuotaMb  uint
	searchQuotaMb uint

	buckets []string
}

const defaultPort = 8091

func defaultNodeConfig() nodeConfig {
	return nodeConfig{
		image:         "docker.io/library/couchbase:community-6.0.0",
		timeoutSecs:   20,
		dataQuotaMb:   1024,
		indexQuotaMb:  256,
		searchQuotaMb: 256,
		port:          defaultPort,
		username:      "Administrator",
		password:      "password",
	}
}

func postNoAuth(t *testing.T, ip, path string, data url.Values) {
	uri, err := url.ParseRequestURI(fmt.Sprintf("http://%s:%d", ip, defaultPort))
	require.NoError(t, err)
	uri.Path = path

	req, err := http.NewRequest("POST", uri.String(), strings.NewReader(data.Encode()))
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	cli := &http.Client{}
	resp, err := cli.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func setMemoryQuotas(t *testing.T, ip, dataMb, indexMb, searchMb string) {
	data := url.Values{}
	data.Set("memoryQuota", dataMb)
	data.Set("indexMemoryQuota", indexMb)
	data.Set("ftsMemoryQuota", searchMb)

	postNoAuth(t, ip, "pools/default", data)
}

func setPortAndCredentials(t *testing.T, ip, port, username, password string) {
	data := url.Values{}
	data.Set("port", port)
	data.Set("username", username)
	data.Set("password", password)

	postNoAuth(t, ip, "settings/web", data)
}

func wait(t *testing.T, ip string, timeout int) {
	secs := 0
	for {
		resp, err := http.Get(fmt.Sprintf("http://%s:%d", ip, defaultPort))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}

		if timeout > 0 && secs >= timeout {
			t.Fatalf("timed out waiting for node %s for %d seconds; try increasing the timeout using node.Setup(Timeout(#))", ip, timeout)
		}

		time.Sleep(1 * time.Second)
		secs++
	}
}
