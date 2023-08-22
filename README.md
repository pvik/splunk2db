# Splunk2DB

A backend ETL service to run search queries against Splunk via API and load them into DB tables.

An alternative to using the Splunk DB Connect App to extract data from Splunk and load it into a relational database.

## Config 

A valid `config.toml` needs to be provided to the `splunk2db` service.

A sample configuration file (with comments) is provided in the `configs` folder.

The config file is split into three main sections (logging is a fourth section, which will assume sane defaults if not defined in the configuration file)

- `database` - defines connection details to databases
- `splunk` - defines connection details to Splunk API
- `queries` - defines search queries that will be run against a splunk instance (defined above) and load them into a database (defined above). This section also allows you to define what fields from the Splunk results will have to pushed to the database. You have the capability to define fields from splunk search to database columns, and also value translation of values returned from splunk to what has to be ingested into the DB.
	
This service allows us to retrieve data larger than data-limits per API call.

If the `allow-partition`, `start-time` and `end-time` fields are populated, the service can automatically (and recursively) reduce the time-interval to pull data, till the results returned are less than the `max-count` limit that Splunk imposes, and the results will be stitched together.


## Usage 

```json
$ ./splunk2db -conf ./config.toml
{"time":"2023-08-22T15:55:53.237388165-04:00","level":"INFO","msg":"splunk2db service initialized"}
{"time":"2023-08-22T15:55:53.237401575-04:00","level":"INFO","msg":"running query","name":"test-query","from-splunk":"datalake1","to-db":"test-mss"}
{"time":"2023-08-22T15:55:53.237406275-04:00","level":"INFO","msg":"connecting to db","db-name":"test-mss"}
{"time":"2023-08-22T15:55:59.84605865-04:00","level":"INFO","msg":"done processing query","query-name":"test-query","processed-records":15}
```

# Thanks

This application uses:

* [go-splunk-rest](https://github.com/pvik/go-splunk-rest) library to handle interaction with Splunk API
* [gorm](gorm.io) library to handle DB interactions