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
	ip          string
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

	var id, ip string
	if conf.alwaysPull {
		id, ip = pullAndStart(t, conf.image)
	} else {
		id, ip = start(t, conf.image)
	}
	n.containerID = id

	wait(t, ip, conf.timeoutSecs)

	setMemoryQuotas(t, ip,
		strconv.FormatUint(uint64(conf.dataQuotaMb), 10),
		strconv.FormatUint(uint64(conf.indexQuotaMb), 10),
		strconv.FormatUint(uint64(conf.searchQuotaMb), 10))

	setServices(t, ip)

	port := strconv.FormatUint(uint64(conf.port), 10)
	setPortAndCredentials(t, ip, port, conf.username, conf.password)
	n.ip = ip
	n.username = conf.username
	n.password = conf.password
}

func (n *Node) Configure(t *testing.T, opts ...ClusterConfigOption) {
	c := connectToCluster(t, n.ip, n.username, n.password)
	c.configure(t, opts...)
	c.close(t)
}

func (n *Node) Teardown(t *testing.T) {
	if n.containerID != "" {
		stopAndRemove(t, n.containerID)
	}
}

func (n *Node) Host() string {
	return "couchbase://" + n.ip
}

func (n *Node) Username() string {
	return n.username
}

func (n *Node) Password() string {
	return n.password
}

type NodeConfigOption func(*nodeConfig)

func DockerImage(image string) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.image = image
	}
}

func AlwaysPull(pull bool) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.alwaysPull = pull
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

type nodeConfig struct {
	image      string
	alwaysPull bool

	timeoutSecs int

	port     uint
	username string
	password string

	dataQuotaMb   uint
	indexQuotaMb  uint
	searchQuotaMb uint
}

const defaultPort = 8091

func defaultNodeConfig() nodeConfig {
	return nodeConfig{
		image:         "docker.io/library/couchbase:community-6.0.0",
		alwaysPull:    true,
		timeoutSecs:   20,
		dataQuotaMb:   1024,
		indexQuotaMb:  256,
		searchQuotaMb: 256,
		port:          defaultPort,
		username:      "Administrator",
		password:      "password",
	}
}

func setMemoryQuotas(t *testing.T, ip, dataMb, indexMb, searchMb string) {
	data := url.Values{}
	data.Set("memoryQuota", dataMb)
	data.Set("indexMemoryQuota", indexMb)
	data.Set("ftsMemoryQuota", searchMb)

	postNoAuth(t, ip, "pools/default", data)
}

func setServices(t *testing.T, ip string) {
	data := url.Values{}
	data.Set("services", "kv,n1ql,index,fts")

	postNoAuth(t, ip, "node/controller/setupServices", data)
}

func setPortAndCredentials(t *testing.T, ip, port, username, password string) {
	data := url.Values{}
	data.Set("port", port)
	data.Set("username", username)
	data.Set("password", password)

	postNoAuth(t, ip, "settings/web", data)
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
	require.Equal(t, http.StatusOK, resp.StatusCode,
		fmt.Sprintf("expected %d, got %s: %+v", http.StatusOK, resp.Status, *resp))
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
		} else {
			time.Sleep(1 * time.Second)
			secs++
		}
	}
}
