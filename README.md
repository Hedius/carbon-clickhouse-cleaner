# graphite-clickhouse-cleaner
Tool for automatically removing / dropping dead rows from go-graphite / carbon-clickhouse / graphite-clickhouse tables.

Be aware. Dropping metrics will cause them to fully be gone.
So if the name of interface changes, or metrics have dynamic tags like a containerID
that changes on redeployments they 
are going to be deleted! Do not use this if you do not want that!

## Workflow
### Query Paths to delete
Firstly, query all dead metrics from the index tables. Paths are logged at level debug.

```sql
SELECT DISTINCT Path, toDateTime(Max(Version)) AS MaxVersion
				FROM graphite_index
				GROUP BY Path
				HAVING MAX(Version) <= toUInt32(toDateTime('XXXX-XX-XX XX:XX:XX'));
SELECT DISTINCT Path, toDateTime(Max(Version)) AS MaxVersion
				FROM graphite_tagged
				GROUP BY Path
				HAVING MAX(Version) <= toUInt32(toDateTime('XXXX-XX-XX XX:XX:XX'));

```

### Delete Points
Secondly, if one of the both query returned Paths, then a `DELETE FROM` is triggered on the points. DELETE/ALTER do not support external tables etc. Therfore, subqueries are used here again.

```sql
DELETE FROM graphite_ng WHERE Date <= toDate(toDateTime('X')) AND (Path IN (
					SELECT DISTINCT Path
					FROM graphite_index
					GROUP BY Path
					HAVING MAX(Version) <= toUInt32(toDateTime('X'))
				) OR Path IN (
                	SELECT DISTINCT Path
					FROM graphite_tagged
					GROUP BY Path
					HAVING MAX(Version) <= toUInt32(toDateTime('X'))
                ))
```

### Delete Index
Thirdly, `ALTER TABLE ... DELETE` is triggered on the index tables.
Assuming those are smaller the heavy DELETE is used here. (Rewrites the whole table.)
Therefore, run the tool like once per week etc...
Reason why this is done is that the value tables do not trigger merges that often.

```sql
ALTER TABLE %s %s DELETE WHERE Date <= toDate(?)
AND Path IN (
    SELECT DISTINCT Path
	FROM %s
	GROUP BY Path
	HAVING MAX(Version) <= toUInt32(toDateTime('X'))
)
```

#### Unused Direct Point delete
If you have points without any index you can try this query to delete them:
```sql
DELETE FROM %s %s WHERE Date <= toDate('X') AND Path IN (
					SELECT DISTINCT Path
					FROM %s
					GROUP BY Path
					HAVING MAX(Timestamp) <= toUInt32(toDateTime('x'))
)
```

# Deployment
## Params
```
Usage of graphite-clickhouse-cleaner:
  -check-config
    	Check config and exit
  -config string
    	Path to config file (default "graphite-clickhouse-cleaner.conf")
  -loglevel string
    	Log level (default "debug")
  -one-shot
        Only run once and not in a loop.
  -dry-run
        Do not execute any DELETEs.
  -version
    	Print version

```

## Config File
The file has to be created manually.

```
[common]
# Max age of plain series before they get dropped. 14d by default.
max-age-plain = "336h"
# Max age of tagged series before they get dropped. 14d by default.
max-age-tagged = "336h"
# Check interval. 7d by default.
loop-interval = "168h"

[clickhouse]
connection-string = "tcp://localhost:9000?username=default&password=&database=default"
value-table = "graphite"
index-table = "graphite_index"
tagged-table = "graphite_tagged"
# Optional -> Cluster (Adds ON CLUSTER X)
# cluster = ""
```
