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

// EnableLogging - disabled by default, but useful is something isn't working.
func EnableLogging() {
	loggingEnabled = true
}

// Node provides access to node-level configuration and operations.
type Node struct {
	containerID string
	ip          string
	username    string
	password    string

	indexTimeout int
}

// NewNode creates a new Node struct.
func NewNode() *Node {
	return &Node{}
}

// Setup creates a new Couchbase node with optional specified configuration.
// Configuration includes the Docker image to use, node memory quotas, and node credentials.
func (n *Node) Setup(t *testing.T, opts ...NodeConfigOption) *Node {
	conf := defaultNodeConfig()
	for _, opt := range opts {
		opt(&conf)
	}
	n.indexTimeout = conf.indexTimeoutSecs

	var id, ip string
	if conf.alwaysPull {
		id, ip = pullAndStart(t, conf.image)
	} else {
		id, ip = start(t, conf.image)
	}
	n.containerID = id

	waitForNode(t, ip, conf.timeoutSecs)
	logf(t, "Container '%s' has started", id)

	setMemoryQuotas(t, ip,
		strconv.FormatUint(uint64(conf.dataQuotaMb), 10),
		strconv.FormatUint(uint64(conf.indexQuotaMb), 10),
		strconv.FormatUint(uint64(conf.searchQuotaMb), 10))
	logf(t, "Memory quotas set [data %d, index %d, search %d]", conf.dataQuotaMb, conf.indexQuotaMb, conf.searchQuotaMb)

	setServices(t, ip)
	logf(t, "Enabled all services")

	port := strconv.FormatUint(uint64(conf.port), 10)
	setPortAndCredentials(t, ip, port, conf.username, conf.password)
	n.ip = ip
	n.username = conf.username
	n.password = conf.password
	logf(t, "Credentials and port set [port %d, username '%s', password '%s']", conf.port, conf.username, conf.password)

	return n
}

// Configure buckets and indexes.
func (n *Node) Configure(t *testing.T, opts ...ClusterConfigOption) {
	c := &Cluster{
		ip:       n.ip,
		username: n.username,
		password: n.password,
	}
	c.configure(t, opts...)
}

// WaitForIndexing to be complete on the specified index.
func (n *Node) WaitForIndexing(t *testing.T, name string, atLeast int) {
	host := fmt.Sprintf("%s:%d", n.ip, defaultPort)

	secs := 0

	var total int
	for {
		var ok bool
		if total, ok = indexDocCount(t, host, n.username, n.password, name); !ok {
			logf(t, "Index '%s' not ready yet", name)
		} else if total < atLeast {
			logf(t, "Index '%s' still only has %d documents, but waiting for %d", name, total, atLeast)
		} else {
			break
		}

		if n.indexTimeout > 0 && secs >= n.indexTimeout {
			t.Fatalf("timed out waiting for indexing of index '%s' for %d seconds; try increasing the timeout using IndexTimeout(#)", name, n.indexTimeout)
		} else {
			time.Sleep(time.Second)
			secs++
		}
	}
	logf(t, "Index '%s' has %d docs", name, total)

	for {
		indexed := sourceDocCount(t, host, n.username, n.password, name)
		perc := 0
		if indexed > 0 {
			perc = total / indexed * 100
		}
		logf(t, "%d%% indexed", perc)

		if indexed >= total {
			break
		}

		if n.indexTimeout > 0 && secs >= n.indexTimeout {
			t.Fatalf("timed out waiting for indexing of index '%s' for %d seconds; try increasing the timeout using IndexTimeout(#)", name, n.indexTimeout)
		} else {
			time.Sleep(time.Second)
			secs++
		}
	}
}

// Teardown the node destroys the container.
func (n *Node) Teardown(t *testing.T) {
	if n.containerID != "" {
		stopAndRemove(t, n.containerID)
		logf(t, "Removed container '%s'", n.containerID)
	}
}

// Host returns connection string that can be used with gocb.Connect().
func (n *Node) Host() string {
	return "couchbase://" + n.ip
}

// Username of the node.
func (n *Node) Username() string {
	return n.username
}

// Password of the node.
func (n *Node) Password() string {
	return n.password
}

// NodeConfigOption functions can be passed to Node.Setup() to configure its creation.
type NodeConfigOption func(*nodeConfig)

// DockerImage configures the Couchbase image.
func DockerImage(image string) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.image = image
	}
}

// AlwaysPull is set to true by default.
// When true, Node.Setup() will attempt to pull the latest image specified by DockerImage().
// When false, Node.Setup() will use the locally available image specified by DockerImage().
// Setting this to true allows the library to be used without a network connection.
func AlwaysPull(pull bool) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.alwaysPull = pull
	}
}

// Timeout configures a time in seconds to wait for a node to be available.
func Timeout(secs int) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.timeoutSecs = secs
	}
}

// Port configures the port.
func Port(port uint) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.port = port
	}
}

// MemoryQuotas configures the data, index and search memory quotas in megabytes.
func MemoryQuotas(dataMb, indexMb, searchMb uint) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.dataQuotaMb = dataMb
		conf.indexQuotaMb = indexMb
		conf.searchQuotaMb = searchMb
	}
}

// Credentials configures the username and password for the node.
// These can be retrieved again with Node.Username() and Node.Password().
func Credentials(username, password string) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.username = username
		conf.password = password
	}
}

// IndexTimeout configures a time in seconds to wait for indexing to complete.
func IndexTimeout(secs int) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.indexTimeoutSecs = secs
	}
}

type nodeConfig struct {
	image      string
	alwaysPull bool

	timeoutSecs      int
	indexTimeoutSecs int

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
		image:            "docker.io/library/couchbase:community-6.0.0",
		alwaysPull:       true,
		timeoutSecs:      20,
		indexTimeoutSecs: 20,
		dataQuotaMb:      1024,
		indexQuotaMb:     256,
		searchQuotaMb:    256,
		port:             defaultPort,
		username:         "Administrator",
		password:         "password",
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

func waitForNode(t *testing.T, ip string, timeout int) {
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

var loggingEnabled bool

func logf(t *testing.T, format string, args ...interface{}) {
	if loggingEnabled {
		t.Logf(format, args...)
	}
}
