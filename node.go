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

	wait(t, ip, defaultPort, conf.timeout)

	port := strconv.FormatUint(uint64(conf.port), 10)
	setPortAndCredentials(t, ip, port, conf.username, conf.password)
	n.host = fmt.Sprintf("couchbase://%s:%d", ip, conf.port)
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
		conf.timeout = secs
	}
}

func Port(port uint) NodeConfigOption {
	return func(conf *nodeConfig) {
		conf.port = port
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
	image    string
	timeout  int
	port     uint
	username string
	password string
	buckets  []string
}

const defaultPort = 8091

func defaultNodeConfig() nodeConfig {
	return nodeConfig{
		image:    "docker.io/library/couchbase:community-6.0.0",
		timeout:  20,
		port:     defaultPort,
		username: "Administrator",
		password: "password",
	}
}

func setPortAndCredentials(t *testing.T, ip, port, username, password string) {
	uri, err := url.ParseRequestURI(fmt.Sprintf("http://%s:%d", ip, defaultPort))
	require.NoError(t, err)
	uri.Path = "settings/web"

	data := url.Values{}
	data.Set("port", port)
	data.Set("username", username)
	data.Set("password", password)

	req, err := http.NewRequest("POST", uri.String(), strings.NewReader(data.Encode()))
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	cli := &http.Client{}
	resp, err := cli.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func wait(t *testing.T, ip string, port uint, timeout int) {
	secs := 0
	for {
		resp, err := http.Get(fmt.Sprintf("http://%s:%d", ip, port))
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
