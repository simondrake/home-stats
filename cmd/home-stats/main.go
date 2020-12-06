package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/simondrake/home-stats/internal/config"
	dbpkg "github.com/simondrake/home-stats/pkg/db"
	hivepkg "github.com/simondrake/home-stats/pkg/hive"
	weatherpkg "github.com/simondrake/home-stats/pkg/weather"
)

func main() {
	conf, err := config.New("settings.json")
	if err != nil {
		log.Fatalf("unable to initialise config: %+v", err)
	}

	hive := hivepkg.New(hivepkg.Config{
		Username: conf.Thermostat.Username,
		Password: conf.Thermostat.Password,
	}, &http.Client{})

	weather := weatherpkg.New(weatherpkg.Config{
		City:    conf.Weather.City,
		Country: conf.Weather.Country,
		APIKey:  conf.Weather.APIKey,
		Units:   conf.Weather.Units,
	}, nil)

	db := dbpkg.New(dbpkg.Config{
		URI:      conf.Database.URI,
		Username: conf.Database.Username,
		Password: conf.Database.Password,
		Database: conf.Database.Database,
	})

	thermostatInterval, err := time.ParseDuration(conf.Thermostat.Interval)
	if err != nil {
		log.Fatalf("unable to parse thermostat interval: %+v", err)
	}

	weatherInterval, err := time.ParseDuration(conf.Weather.Interval)
	if err != nil {
		log.Fatalf("unable to parse weather interval: %+v", err)
	}

	fmt.Printf(`Config Values set
  Thermostat Enabled: %t
  Thermostat Interval: %s
  AutoBoost Enabled: %t
  AutoBoost Min Temperature: %f
  Weather Enabled: %t
  Weather Interval: %s

`,
		conf.Thermostat.Enabled, conf.Thermostat.Interval, conf.Thermostat.AutoBoost.Enabled, conf.Thermostat.AutoBoost.MinTemperature, conf.Weather.Enabled, conf.Weather.Interval)

	it := time.NewTicker(thermostatInterval)
	wt := time.NewTicker(weatherInterval)

	for {
		select {
		case <-it.C:
			if conf.Thermostat.Enabled {
				log.Println("Getting theromostat statistics")

				if err := hive.GenerateToken(); err != nil {
					fmt.Printf("error generating token: %+v", err)
					log.Fatalf("error generating token: %+v", err)
				}

				thermostatTemp, err := hive.GetTempForNode(conf.Thermostat.ThermostatID)
				if err != nil {
					fmt.Printf("error getting temp for thermostat (%s): %+v", conf.Thermostat.ThermostatID, err)
					log.Fatalf("error getting temp for thermostat (%s): %+v", conf.Thermostat.ThermostatID, err)
				}

				db.Write(context.Background(), dbpkg.WriteRequest{
					Measurement: "thermostat",
					Tags: map[string]string{
						"unit": "temperature",
					},
					Fields: map[string]interface{}{
						"current": thermostatTemp,
					},
					Timestamp: time.Now(),
				})

				// If AutoBoost is enabled, we check if the minimum temperature has been met.
				// If it has we boost the heating
				if conf.Thermostat.AutoBoost.Enabled && thermostatTemp <= conf.Thermostat.AutoBoost.MinTemperature {
					log.Println("Boosting heating")

					err := hive.BoostHeating(
						conf.Thermostat.ThermostatID,
						conf.Thermostat.AutoBoost.TargetDuration,
						conf.Thermostat.AutoBoost.TargetTemperature,
					)
					if err != nil {
						fmt.Printf("error boosting the heating: %+v", err)
						log.Fatalf("error boosting the heating: %+v", err)
					}
				}
			}

		case <-wt.C:
			if conf.Weather.Enabled {
				log.Println("Getting weather statistics")

				currentWeather, err := weather.GetCurrentWeather()
				if err != nil {
					fmt.Printf("error getting current weather: %+v \n", err)
					log.Fatalf("error getting current weather: %+v", err)
				}

				db.Write(context.Background(), dbpkg.WriteRequest{
					Measurement: "weather",
					Tags: map[string]string{
						"unit": "temperature",
					},
					Fields: map[string]interface{}{
						"current": currentWeather.Main.Temperature,
					},
					Timestamp: time.Now(),
				})
			}

		}
	}
}
