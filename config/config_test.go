package config

import (
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfigRead(t *testing.T) {
	input := `
[common]
max-age-plain = "1h"
max-age-tagged = "3h"
loop-interval = "2h"

[clickhouse]
connection-string = "bla"
value-table = "test"
index-table = "test1"
tagged-table = "test2"
`
	var cfg Config
	err := toml.Unmarshal([]byte(input), &cfg)
	require.NoError(t, err)

	expected := New()

	expected.Common = Common{
		MaxAgePlain:  Duration(1 * time.Hour),
		MaxAgeTagged: Duration(3 * time.Hour),
		LoopInterval: Duration(2 * time.Hour),
	}

	expected.ClickHouse = ClickHouse{
		ConnectionString: "bla",
		ValueTable:       "test",
		IndexTable:       "test1",
		TaggedTable:      "test2",
	}

	assert.Equal(t, expected, &cfg)
}
