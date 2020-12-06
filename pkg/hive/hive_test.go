package hive_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/simondrake/home-stats/pkg/hive"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	req      *http.Request
	response *http.Response
	err      error
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	m.req = req
	return m.response, m.err
}

func TestGenerateToken(t *testing.T) {
	t.Skip("Find a way of testing cognitosrp")

	t.Run("should return an error when the request fails", func(t *testing.T) {
		a := assert.New(t)

		mc := &mockClient{response: nil, err: errors.New("something went wrong")}

		h := hive.New(hive.Config{}, mc)

		a.EqualError(h.GenerateToken(), "error requesting login token: something went wrong")
	})
	t.Run("should return an error if no token is returned", func(t *testing.T) {
		a := assert.New(t)

		r := ioutil.NopCloser(bytes.NewReader([]byte(`{"message": "everythings fine, but no token"}`)))
		mc := &mockClient{response: &http.Response{StatusCode: http.StatusOK, Body: r}}

		h := hive.New(hive.Config{}, mc)

		a.EqualError(h.GenerateToken(), "no token returned")
	})
	t.Run("should set the token when returned", func(t *testing.T) {
		a := assert.New(t)

		r := ioutil.NopCloser(bytes.NewReader([]byte(`{"token": "a-special-token"}`)))
		mc := &mockClient{response: &http.Response{StatusCode: http.StatusOK, Body: r}}

		h := hive.New(hive.Config{}, mc)

		a.NoError(h.GenerateToken())

		// Call GetTempForNode so a new request is made to Do with a request
		// containing the token generated in the last step

		// There will be an error because we don't return any node information
		_, _ = h.GetTempForNode("test-node")

		a.Equal("a-special-token", mc.req.Header.Get("X-Omnia-Access-Token"))
	})
}
