package weather_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/simondrake/home-stats/pkg/weather"
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

func TestGetCurrentWeather(t *testing.T) {
	t.Run("should return an error and an empty CurrentWeather struct", func(t *testing.T) {
		mc := &mockClient{response: nil, err: errors.New("something went wrong")}

		w := weather.New(weather.Config{}, mc)

		cw, err := w.GetCurrentWeather()

		assert.EqualError(t, err, "error requesting node information: something went wrong")
		assert.Empty(t, cw)
	})

	t.Run("should marshal API response into a CurrentWeather struct", func(t *testing.T) {
		r := ioutil.NopCloser(bytes.NewReader([]byte(`{"base": "base-test","visibility": 100}`)))
		mc := &mockClient{response: &http.Response{StatusCode: http.StatusOK, Body: r}}

		w := weather.New(
			weather.Config{
				City:    "London",
				Country: "GB",
				APIKey:  "MockKey",
			},
			mc,
		)

		cw, err := w.GetCurrentWeather()
		assert.NoError(t, err)

		assert.Equal(t, "base-test", cw.Base)
		assert.Equal(t, int32(100), cw.Visibility)
	})

	t.Run("should set the units query parameter when Units is set", func(t *testing.T) {
		r := ioutil.NopCloser(bytes.NewReader([]byte(`{"base": "base-test","visibility": 100}`)))
		mc := &mockClient{response: &http.Response{StatusCode: http.StatusOK, Body: r}}

		w := weather.New(
			weather.Config{
				City:    "London",
				Country: "GB",
				APIKey:  "MockKey",
				Units:   "metric",
			},
			mc,
		)

		cw, err := w.GetCurrentWeather()

		assert.NoError(t, err)
		assert.NotEmpty(t, cw)

		expectedURL := fmt.Sprintf("%s?q=%s,%s&appid=%s&units=%s", "https://api.openweathermap.org/data/2.5/weather", w.City, w.Country, w.APIKey, w.Units)
		assert.Equal(t, expectedURL, mc.req.URL.String())
	})
}

// Test that is able to mock the HTTP request in
// GetCurrentWeather by refactoring the function to
// accept a URL in the struct.

// We're not going to use this strategy, but I'll leave it here so it can
// be comitted in case I want to reference it later.

// TODO - remove this in the next commit

// func TestGetCurrentWeather(t *testing.T) {
// 	w := &weather.Weather{
// 		weather.Config{
// 			City:    "London",
// 			Country: "GB",
// 			APIKey:  "MockKey",
// 		},
// 	}
//
// 	mockResponse := func(w http.ResponseWriter, r *http.Request) {
// 		_, _ = w.Write([]byte(`{"base": "base-test","visibility": 100}`))
// 	}
//
// 	handler := http.NewServeMux()
// 	handler.HandleFunc("/", mockResponse)
//
// 	srv := httptest.NewServer(handler)
// 	defer srv.Close()
//
// 	w.Endpoint = srv.URL
//
// 	cw, err := w.GetCurrentWeather()
// 	assert.NoError(t, err)
//
// 	assert.Equal(t, "base-test", cw.Base)
// 	assert.Equal(t, int32(100), cw.Visibility)
// }
