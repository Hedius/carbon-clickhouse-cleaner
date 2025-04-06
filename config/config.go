package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Duration time.Duration

type Common struct {
	MaxAgePlain  Duration `toml:"max-age-plain"  json:"max-age-plain"  comment:"Max age of plain series before they get dropped. 14d by default."`
	MaxAgeTagged Duration `toml:"max-age-tagged" json:"max-age-tagged" comment:"Max age of tag before they get dropped. 14d by default."`
	LoopInterval Duration `toml:"loop-interval"  json:"loop-interval"  comment:"Check interval. 1h by default."`
}

type ClickHouse struct {
	ConnectionString string `toml:"connection-string" json:"connection-string" comment:"ClickHouse connection string"`
	ValueTable       string `toml:"value-table"       json:"value-table"       comment:"Name of value table. graphite by default."`
	IndexTable       string `toml:"index-table"       json:"index-table"       comment:"Name of index table in graphite_index by default."`
	TaggedTable      string `toml:"tagged-table"      json:"tagged-table"      comment:"Name of tagged table in graphite_tagged by default."`
}

// Config is the main config struct.
type Config struct {
	Common     Common     `toml:"common"     json:"common"`
	ClickHouse ClickHouse `toml:"clickhouse" json:"clickhouse"`
}

// New returns *Config with default values.
func New() *Config {
	cfg := &Config{
		Common: Common{
			MaxAgePlain:  Duration(2 * 168 * time.Hour),
			MaxAgeTagged: Duration(2 * 168 * time.Hour),
			LoopInterval: Duration(1 * time.Hour),
		},
		ClickHouse: ClickHouse{
			ConnectionString: "tcp://localhost:9000?username=default&password=&database=default",
			ValueTable:       "graphite",
			IndexTable:       "graphite_index",
			TaggedTable:      "graphite_tagged",
		},
	}
	return cfg
}

func (d *Duration) UnmarshalText(b []byte) error {
	x, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(x)
	return nil
}

// PrintDefaultConfig prints the default config to stdout
func PrintDefaultConfig() {
	cfg := New()
	b, err := toml.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

// ReadConfig parses the config at the given path.
func ReadConfig(filename string) (*Config, error) {
	var err error
	var body []byte
	if filename == "" {
		return nil, errors.New("filename is required")
	}
	body, err = os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, errors.New("config file is empty")
	}
	var cfg Config
	err = toml.Unmarshal(body, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
