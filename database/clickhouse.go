package database

import (
	"database/sql"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	log "github.com/sirupsen/logrus"
	"time"
)

type ClickHouse struct {
	ConnectionString string
	ValueTable       string
	IndexTable       string
	TaggedTable      string

	con *sql.DB
}

// Open gets a CH connection pool
func (ch *ClickHouse) Open() (*sql.DB, error) {
	db, err := sql.Open("clickhouse", ch.ConnectionString)
	ch.con = db
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, err
}

func (ch *ClickHouse) Close() {
	if ch.con != nil {
		err := ch.con.Close()
		if err != nil {
			log.Error("Error closing ClickHouse connection: ", err)
		}
		ch.con = nil
	}
}

func (ch *ClickHouse) GetPathsToDelete(table string, maxTimeStamp time.Time) ([]string, error) {
	query := fmt.Sprintf(
		`SELECT DISTINCT Path, toDateTime(Max(Version)) AS MaxVersion
				FROM %s
				GROUP BY Path
				HAVING MAX(Version) <= toUInt32(?)`, table)
	rows, err := ch.con.Query(query, maxTimeStamp)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()
	var paths []string
	for rows.Next() {
		var path string
		var maxVersion time.Time
		if err := rows.Scan(&path, &maxVersion); err != nil {
			return nil, err
		}
		paths = append(paths, path)
		log.Debugf("Path %v can be deleted. Last seen at %v", path, maxVersion)
	}
	log.Infof("%v paths/metrics can be deleted from %v", len(paths), table)
	return paths, nil
}

func (ch *ClickHouse) DeletePoints(maxTimeStampPlain time.Time, maxTimeStampTagged time.Time) error {
	var minTimeStamp time.Time
	if maxTimeStampPlain.After(maxTimeStampTagged) {
		minTimeStamp = maxTimeStampTagged
	} else {
		minTimeStamp = maxTimeStampTagged
	}
	query := fmt.Sprintf(
		`DELETE FROM %s WHERE Date <= toDate(?) AND Path IN (
					SELECT DISTINCT Path
					FROM %s
					GROUP BY Path
					HAVING MAX(Version) <= toUInt32(?)
				) OR Path IN (
                	SELECT DISTINCT Path
					FROM %s
					GROUP BY Path
					HAVING MAX(Version) <= toUInt32(?)
				)`, ch.ValueTable, ch.IndexTable, ch.TaggedTable)
	log.Infof("[%v] Triggering delete", ch.ValueTable)
	_, err := ch.con.Exec(query, minTimeStamp, maxTimeStampPlain, maxTimeStampTagged)
	if err != nil {
		log.Errorf("[%v] Error while deleting points: %v", ch.ValueTable, err)
	}
	log.Infof("[%v] Delete executed", ch.ValueTable)
	return err
}

func (ch *ClickHouse) DeletePaths(table string, maxTimeStamp time.Time) error {
	query := fmt.Sprintf(
		`ALTER TABLE %s DELETE WHERE Date <= toDate(?)
                AND Path IN (
					SELECT DISTINCT Path
					FROM %s
					GROUP BY Path
					HAVING MAX(Version) <= toUInt32(?)
				)`, table, table)
	log.Infof("[%v] Triggering delete", table)
	_, err := ch.con.Exec(query, maxTimeStamp, maxTimeStamp)
	if err != nil {
		log.Errorf("[%v] Error while deleting index: %v", table, err)
		return err
	}
	log.Infof("[%v] Delete executed", table)
	return err
}
