package main

import (
	"flag"
	"fmt"
	"graphite-clickhouse-cleaner/config"
	"graphite-clickhouse-cleaner/database"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const Version = "0.1.0"

// 1. config reading
// 2. setup logging
// 3. Loop
//  1. query untagged
//  2. query tagged
//  3. trigger value delete.
//  4. delete index/tagged

type Cleaner struct {
	cfg   *config.Config
	ch     database.ClickHouse
}

func (c *Cleaner) Clean() error {
	log.Info("Starting cleanup...")
	_, err := c.ch.Open()
	if err != nil {
		log.Error("Error opening ClickHouse connection: ", err)
		return err
	}
	defer c.ch.Close()

	now := time.Now()
	timeStampPlain := now.Add(-time.Duration(c.cfg.Common.MaxAgePlain))
	timeStampTagged := now.Add(-time.Duration(c.cfg.Common.MaxAgeTagged))

	// 1. Get obsolete paths from index/tagged
	obsoletePlainPaths, err := c.ch.GetPathsToDelete(c.ch.IndexTable, timeStampPlain)
	if err != nil {
		return err
	}
	obsoleteTaggedPaths, err := c.ch.GetPathsToDelete(c.ch.TaggedTable, timeStampTagged)
	if err != nil {
		return err
	}

	if len(obsoletePlainPaths) == 0 && len(obsoleteTaggedPaths) == 0 {
		return nil
	}

	// 2. Trigger a DELETE for the points
	err = c.ch.DeletePoints(timeStampPlain, timeStampTagged)
	if err != nil {
		return err
	}

	// 3. Delete the paths from index/tagged tables with alter statements.
	if len(obsoletePlainPaths) > 0 {
		err = c.ch.DeletePaths(c.ch.IndexTable, timeStampPlain)
		if err != nil {
			return err
		}
	}
	if len(obsoleteTaggedPaths) > 0 {
		err = c.ch.DeletePaths(c.ch.TaggedTable, timeStampTagged)
		if err != nil {
			return err
		}
	}

	log.Info("Cleanup finished")
	return nil
}

func main() {
	var err error
	configFile := flag.String("config", "graphite-clickhouse-cleaner.conf", "path to config file")
	checkConfig := flag.Bool("check-config", false, "check config file and exit")
	logLevel := flag.String("loglevel", "debug", "log level")
	printVersion := flag.Bool("version", false, "print version and exit")
	oneShot := flag.Bool("one-shot", false, "only run once if set")
	dryRun := flag.Bool("dry-run", true, "enable debug mode (Do not run any delete)")

	flag.Parse()

	lvl, err := log.ParseLevel(*logLevel)
	log.SetLevel(lvl)

	if *printVersion {
		fmt.Print(Version)
		return
	}

	cfg, err := config.ReadConfig(*configFile)

	if err != nil {
		log.Fatal(err)
	}

	if *checkConfig {
		return
	}

	cleaner := &Cleaner{
		cfg:   cfg,
		ch: database.ClickHouse{
			ConnectionString: cfg.ClickHouse.ConnectionString,
			ValueTable:       cfg.ClickHouse.ValueTable,
			IndexTable:       cfg.ClickHouse.IndexTable,
			TaggedTable:      cfg.ClickHouse.TaggedTable,
			Cluster:          cfg.ClickHouse.Cluster,
			DryRun: 		  *dryRun,
		},
	}

	for {
		err = cleaner.Clean()
		if err != nil {
			log.Error("Error cleaning clickhouse: ", err)
			if *oneShot {
				os.Exit(1)
			}
		}

		if *oneShot {
			// abort if we one shot
			break
		}

		log.Info("Sleeping for ", time.Duration(cfg.Common.LoopInterval))
		time.Sleep(time.Duration(cfg.Common.LoopInterval))
	}

}
