package main

import (
	"fmt"

	"github.com/pvik/splunk2db/internal/config"
	"github.com/pvik/splunk2db/internal/service"
	"github.com/pvik/splunk2db/pkg/db"
	"gorm.io/gorm"

	splunk "github.com/pvik/go-splunk-rest"

	log "log/slog"
)

func init() {
	// Initialize config file
	// Setup Logging
	service.InitService()
}

func main() {
	defer service.Shutdown()

	dbConnMap := make(map[string]*gorm.DB)

	for qName, query := range config.AppConf.Queries {
		log.Info("running query",
			"name", qName,
			"from-splunk", query.SplunkName,
			"to-db", query.DBName,
		)

		log.Debug("query details", "query", query)

		_, splunkConnExists := config.AppConf.Splunk[query.SplunkName]
		if !splunkConnExists {
			log.Error("splunk connection not configured for query, skipping query",
				"query-name", qName,
				"invalid-splunk-name", query.SplunkName,
			)
			continue
		}

		_, dbConnExists := config.AppConf.DB[query.DBName]
		if !dbConnExists {
			log.Error("database connection not configured for query, skipping query",
				"query-name", qName,
				"invalid-db-name", query.DBName,
			)
			continue
		}

		if dbConnMap[query.DBName] == nil {
			log.Info("connecting to db",
				"db-name", query.DBName,
			)
			conn, err := db.Connect(
				config.AppConf.DB[query.DBName].DBType,
				config.AppConf.DB[query.DBName].Host,
				config.AppConf.DB[query.DBName].Port,
				config.AppConf.DB[query.DBName].DBName,
				config.AppConf.DB[query.DBName].Username,
				config.AppConf.DB[query.DBName].Password,
				config.AppConf.DB[query.DBName].ExtraParam,
				"info",
			)
			if err != nil {
				log.Error("unable to connect to db, skipping query",
					"query-name", qName,
					"db-name", query.DBName,
					"err", err,
				)
				continue
			}

			dbConnMap[query.DBName] = conn
		}

		searchOptions := splunk.SearchOptions{
			AllowPartition: query.AllowPartition,
		}

		if config.AppConf.Splunk[query.SplunkName].MaxCount != 0 {
			searchOptions.MaxCount = config.AppConf.Splunk[query.SplunkName].MaxCount
		}

		if query.StartTime != "" {
			timeVal, isValidTime, err := parseTimeString(query.StartTime)
			if err != nil {
				log.Warn("invalid start time",
					"query-name", qName,
					"start-time", query.StartTime,
					"error", err,
				)
				continue
			}

			if isValidTime {
				searchOptions.EarliestTime = timeVal
				searchOptions.UseEarliestTime = true

				log.Debug("search limit start", "time", searchOptions.EarliestTime)
			}
		}

		if query.EndTime != "" {
			timeVal, isValidTime, err := parseTimeString(query.EndTime)
			if err != nil {
				log.Warn("invalid start time",
					"query-name", qName,
					"end-time", query.EndTime,
					"error", err,
				)
				continue
			}

			if isValidTime {
				searchOptions.LatestTime = timeVal
				searchOptions.UseLatestTime = true

				log.Debug("search limit end", "time", searchOptions.EarliestTime)
			}
		}

		// Search Splunk with query
		recs, err := config.AppConf.Splunk[query.SplunkName].Search(query.Search, searchOptions)
		if err != nil {
			log.Error("error searching splunk", "query", qName, "error", err)
		}

		// Convert Splunk results to DB records
		processedRecs := processRecords(recs, query.IncludeFields, query.SplunkFields)

		recordsCount := 0
		for _, r := range processedRecs {
			log.Debug("process record to db", "query", qName, "fields", r)

			recExistsInDB := false

			// Check if record exists with primary key,
			//   iff DBPrimaryKeyCol is defined for query
			if query.DBPrimaryKeyCol != "" {
				dbRec := make(map[string]interface{})
				dbConnMap[query.DBName].Table(query.DBTableName).
					Where(fmt.Sprintf("%s = ?", query.DBPrimaryKeyCol), r[query.DBPrimaryKeyCol]).
					Scan(&dbRec)

				_, recExistsInDB = dbRec[query.DBPrimaryKeyCol]
			}

			if recExistsInDB {
				log.Debug("update record in DB", query.DBPrimaryKeyCol, r[query.DBPrimaryKeyCol])
				// update record
				res := dbConnMap[query.DBName].Table(query.DBTableName).
					Where(fmt.Sprintf("%s = ?", query.DBPrimaryKeyCol), r[query.DBPrimaryKeyCol]).
					Updates(r)

				if res.Error != nil {
					log.Error("unable to update record", query.DBPrimaryKeyCol, r[query.DBPrimaryKeyCol], "err", res.Error)
				}
			} else {
				log.Debug("insert record to DB", query.DBPrimaryKeyCol, r[query.DBPrimaryKeyCol])
				// insert record
				res := dbConnMap[query.DBName].Table(query.DBTableName).
					Create(r)
				if res.Error != nil {
					log.Error("unable to insert record", "err", res.Error)
				}

			}
			recordsCount = recordsCount + 1

			if recordsCount%500 == 0 {
				log.Info("processing query",
					"query-name", qName,
					"processed-count", recordsCount,
				)
			}
		}
		log.Info("done processing query", "query-name", qName, "processed-records", recordsCount)

	}
}

func processRecords(records []map[string]interface{},
	fields []string,
	splunkFieldTranslations map[string]config.SplunkFieldTranslation) []map[string]interface{} {
	processedRecords := make([]map[string]interface{}, 0, len(records))

	for _, rec := range records {
		processedRec := make(map[string]interface{})

		for _, field := range fields {
			// check if splunk field translation exists
			splunkTranslation, splunkTranslationExists := splunkFieldTranslations[field]
			if splunkTranslationExists {

				recFieldValString := stringValFromInterface(rec[field])
				valueTranslation, valueTranslationExists := splunkTranslation.ValueTranslation[recFieldValString]
				if valueTranslationExists {
					processedRec[splunkTranslation.DBColumnName] = valueTranslation
				} else {
					processedRec[splunkTranslation.DBColumnName] = rec[field]
				}
			} else {
				processedRec[field] = rec[field]
			}
		}

		processedRecords = append(processedRecords, processedRec)
	}

	return processedRecords
}
