package db

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Config struct {
	// URI is the host:port combination (e.g. http://localhost:8086)
	URI string
	// Username is the database username
	Username string
	// Password is the database password
	Password string
	// Database is the database name
	Database string
}

type DB struct {
	Config
}

type WriteRequest struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Timestamp   time.Time
}

func New(c Config) *DB {
	return &DB{
		Config: c,
	}
}

func (d *DB) Write(ctx context.Context, wr WriteRequest) error {
	c := influxdb2.NewClient(d.URI, fmt.Sprintf("%s:%s", d.Username, d.Password))
	defer c.Close()

	writeAPI := c.WriteAPIBlocking("", d.Database)

	p := influxdb2.NewPoint(wr.Measurement, wr.Tags, wr.Fields, wr.Timestamp)

	return writeAPI.WritePoint(ctx, p)
}
