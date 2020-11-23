package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const openWeatherMapEndpoint = "https://api.openweathermap.org/data/2.5/weather"

type Config struct {
	City    string
	Country string
	APIKey  string
	Units   string
}

type Weather struct {
	Config
}

func New(c Config) *Weather {
	return &Weather{
		Config: c,
	}
}

type CurrentWeather struct {
	Coord             Coord              `json:"coord,omitempty"`
	WeatherConditions []WeatherCondition `json:"weather,omitempty"`
	Base              string             `json:"base,omitempty"`
	Main              Main               `json:"main,omitempty"`
	Visibility        int32              `json:"visibility,omitempty"`
	Wind              Wind               `json:"wind,omitempty"`
	Clouds            Clouds             `json:"clouds,omitempty"`
	// Timestamp is the time of the data calculation, unix, UTC
	Timestamp int64 `json:"dt,omitempty"`
	Sys       Sys   `json:"sys,omitempty"`
	// Timezone is the shift in seconds from UTC
	Timezone int64 `json:"timezone,omitempty"`
	// ID is the City ID
	ID int32 `json:"id,omitempty"`
	// Name is the City name
	Name string `json:"name,omitempty"`
	// Cod is an internal parameter to OpenWeatherMap
	Cod int32 `json:"cod,omitempty"`
}

type Coord struct {
	// Lon is the citys Longitude
	Lon string `json:"lon,omitempty"`
	// Lat is the citys Latitude
	Lat string `json:"lat,omitempty"`
}

type WeatherCondition struct {
	// ID is the weather condition id
	ID string `json:"id,omitempty"`
	// Main is the group of weather parameters (Rain, Snow, Extreme etc)
	Main string `json:"main,omitempty"`
	// Description is the weather condition within the group
	Description string `json:"description,omitempty"`
	// Icon is the weather icon id
	Icon string `json:"icon,omitempty"`
}

type Main struct {
	// Temperature is the temperature
	Temperature float32 `json:"temp,omitempty"`
	// FeelsLike accounts for the human perception of weather
	FeelsLike float32 `json:"feels_like,omitempty"`
	// Pressure is the Atmospheric pressure
	Pressure int32 `json:"pressure,omitempty"`
	// Humidity is the humidity
	Humidity int32 `json:"humidity,omitempty"`
	// TemperatureMin is the minimum temperature at the moment
	TemperatureMin float32 `json:"temp_min,omitempty"`
	// TemperatureMax is the maximum temperature at the moment
	TemperatureMax float32 `json:"temp_max,omitempty"`
}

type Wind struct {
	// Speed is the wind speed
	Speed float32 `json:"speed,omitempty"`
	// Direction is the wind direction in degrees
	Direction int32 `json:"deg,omitempty"`
	// Gust is the wind gust
	Gust float32 `json:"gust,omitempty"`
}

type Clouds struct {
	// All is the cloudiness
	All int32 `json:"all,omitempty"`
}

type Sys struct {
	// Type is an internal parameter to OpenWeatherMap
	Type int32 `json:"type,omitempty"`
	// ID is an internal parameter to OpenWeatherMap
	ID int32 `json:"id,omitempty"`
	// Country is the Country code
	Country string `json:"country,omitempty"`
	// Sunrise is the sunrise time, unix, UTC
	Sunrise int64 `json:"sunrise,omitempty"`
	// Sunset is the sunset time, unix, UTC
	Sunset int64 `json:"sunset,omitempty"`
}

func (w *Weather) GetCurrentWeather() (CurrentWeather, error) {
	client := &http.Client{}

	endpoint := fmt.Sprintf("%s?q=%s,%s&appid=%s", openWeatherMapEndpoint, w.City, w.Country, w.APIKey)

	if w.Units != "" {
		endpoint += "&units=" + w.Units
	}

	var cw CurrentWeather

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return cw, fmt.Errorf("error creating request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return cw, fmt.Errorf("error requesting node information: %w", err)
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&cw)

	return cw, nil
}
