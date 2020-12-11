package speedtest

import (
	"fmt"
	"net/http"
)

const (
	speedtestEndpoint = "http://speedtest.net/"
	userPath          = "speedtest-config.php"
)

type Config struct{}

type SpeedTest struct {
	httpClient httpClient
	Config
}

type Users struct {
	Users []User `xml:"client"`
}

// User information
type User struct {
	IP  string `xml:"ip,attr"`
	Lat string `xml:"lat,attr"`
	Lon string `xml:"lon,attr"`
	Isp string `xml:"isp,attr"`
}

type ServerList struct {
	Servers []Server `xml:"servers>server"`
}

type Server struct {
	URL      string `xml:"url,attr"`
	Lat      string `xml:"lat,attr"`
	Lon      string `xml:"lon,attr"`
	Name     string `xml:"name,attr"`
	Country  string `xml:"country,attr"`
	Sponsor  string `xml:"sponsor,attr"`
	ID       string `xml:"id,attr"`
	URL2     string `xml:"url2,attr"`
	Host     string `xml:"host,attr"`
	Distance float64
	DLSpeed  float64
	ULSpeed  float64
}

// httpClient implements the Do method, which is the exact
// API of the http.Client's DO function. This helps with testing.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(c Config, client httpClient) *SpeedTest {
	if client == nil {
		client = &http.Client{}
	}

	return &SpeedTest{
		httpClient: client,
		Config:     c,
	}
}

func (s *SpeedTest) Run() error {
	return nil
}

func (s *SpeedTest) getUserInfo() (*User, error) {
	req, err := http.NewRequest("GET", speedtestEndpoint+userPath, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}

	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	defer res.Body.Close()

	return nil, nil
}
