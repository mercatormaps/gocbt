package gocbt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func indexDocCount(t *testing.T, host, username, password, name string) (int, bool) {
	type Response struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}

	resp := Response{}
	getWithAuth(t, fmt.Sprintf("http://%s/_p/fts/api/index/%s/count", host, name), username, password, &resp)
	return resp.Count, resp.Status == "ok"
}

func sourceDocCount(t *testing.T, host, username, password, name string) int {
	type Response struct {
		Count int `json:"docCount"`
	}

	resp := Response{}
	getWithAuth(t, fmt.Sprintf("http://%s/_p/fts/api/stats/sourceStats/%s", host, name), username, password, &resp)
	return resp.Count
}

func getWithAuth(t *testing.T, url, username, password string, v interface{}) {
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	req.SetBasicAuth(username, password)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, v)
	require.NoError(t, err)
}
