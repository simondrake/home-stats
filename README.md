# Home Stats

Home Stats is a small utility that does the following:

* Queries the Hive Heating API on the interval set in `settings.json`
  * Stores the temperature of the thermostat ID, specified in `settings.json`, into Influx.
  * If the temperature is <= the `minTemperature`, specified in `settings.json` it will boost the heating for the duration specified.
* Queries the [OpenWeather](https://openweathermap.org/api) API, gets the temperature and stores it in Influx.

# Attributions

* [cognito-srp](https://github.com/alexrudd/cognito-srp) by [@alexrudd](https://github.com/alexrudd), which was copied and added to this project.

# TODO

* [ ] Add instructions for getting Hive thermostat ID
* [ ] Add Speedtest package
* [x] Write README
* [x] Tests
