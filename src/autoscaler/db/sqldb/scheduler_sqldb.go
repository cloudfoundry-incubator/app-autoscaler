package sqldb

import (
	"code.cloudfoundry.org/lager"
	_ "github.com/lib/pq"

	"autoscaler/db"
	"autoscaler/models"

	"database/sql"
	"strconv"
	"strings"
)

type SchedulerSQLDB struct {
	sqldb  *sql.DB
	logger lager.Logger
	url    string
}

func NewSchedulerSQLDB(url string, logger lager.Logger) (*SchedulerSQLDB, error) {
	sqldb, err := sql.Open(db.PostgresDriverName, url)
	if err != nil {
		logger.Error("failed-open-scheduler-db", err, lager.Data{"url": url})
		return nil, err
	}

	err = sqldb.Ping()
	if err != nil {
		sqldb.Close()
		logger.Error("failed-ping-scheduler-db", err, lager.Data{"url": url})
		return nil, err
	}

	return &SchedulerSQLDB{
		url:    url,
		logger: logger,
		sqldb:  sqldb,
	}, nil
}

func (sdb *SchedulerSQLDB) Close() error {
	err := sdb.sqldb.Close()
	if err != nil {
		sdb.logger.Error("failed-close-scheduler-db", err, lager.Data{"url": sdb.url})
		return err
	}
	return nil
}

func (sdb *SchedulerSQLDB) GetActiveSchedules() (map[string]*models.ActiveSchedule, error) {
	query := "SELECT id, app_id, instance_min_count, instance_max_count, initial_min_instance_count FROM app_scaling_active_schedule"
	rows, err := sdb.sqldb.Query(query)
	if err != nil {
		sdb.logger.Error("failed-get-active-schedules-query", err, lager.Data{"query": query})
		return nil, err
	}
	defer rows.Close()

	schedules := make(map[string]*models.ActiveSchedule)
	var id int64
	var appId string
	var instanceMin, instanceMax int
	minInitial := sql.NullInt64{}
	for rows.Next() {
		if err = rows.Scan(&id, &appId, &instanceMin, &instanceMax, &minInitial); err != nil {
			sdb.logger.Error("failed-get-active-schedules-scan", err)
			return nil, err
		}
		instanceMinInitial := 0
		if minInitial.Valid {
			instanceMinInitial = int(minInitial.Int64)
		}

		schedule := models.ActiveSchedule{
			ScheduleId:         strconv.FormatInt(id, 10),
			InstanceMin:        instanceMin,
			InstanceMax:        instanceMax,
			InstanceMinInitial: instanceMinInitial,
		}
		schedules[appId] = &schedule
	}
	return schedules, nil

}

func (sdb *SchedulerSQLDB) SynchronizeActiveSchedules(appIdMap map[string]bool) error {
	if len(appIdMap) == 0 {
		sdb.logger.Debug("No application exists")
		return nil
	}
	var query string = "DELETE FROM app_scaling_active_schedule WHERE app_id NOT IN("
	var appIdsStr string = ""
	for appId, _ := range appIdMap {
		appIdsStr += "'" + appId + "',"
	}
	appIdsStr = strings.TrimRight(appIdsStr, ",")
	query += appIdsStr + ")"
	_, err := sdb.sqldb.Exec(query)
	if err != nil {
		sdb.logger.Error("Failed to delete active schedules", err, lager.Data{"query": query})
	}

	return err
}
