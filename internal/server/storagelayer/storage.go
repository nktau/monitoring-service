package storagelayer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nktau/monitoring-service/internal/server/config"
	"github.com/nktau/monitoring-service/internal/server/utils"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

type memStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
	logger  *zap.Logger
	config  config.Config
}

type MemStorage interface {
	UpdateCounter(string, int64) error
	UpdateGauge(string, float64) error
	GetCounter(string) (int64, error)
	GetGauge(string) (float64, error)
	GetAll() (map[string]float64, map[string]int64)
	CheckDBConnection() error
	Updates(metric []utils.Metrics) error
}

var ErrMetricNotFound = errors.New("metric not found")

const dataPathDisk string = "disk"
const dataPathDatabase string = "database"

func New(logger *zap.Logger, config config.Config) *memStorage {
	mem := &memStorage{
		Gauge:   map[string]float64{},
		Counter: map[string]int64{},
		logger:  logger,
		config:  config,
	}
	if config.Restore && config.DatabaseDSN != "" {
		mem.logger.Info("mem.readFromDB()")
		mem.readFromDB()
	} else if config.Restore && config.FileStoragePath != "" {
		mem.readFromDisk()
	}

	if config.StoreInterval != 0 && config.DatabaseDSN != "" {
		mem.createDBScheme()
		go mem.storeDataWithStoreInterval(dataPathDatabase)
	} else if config.StoreInterval != 0 && config.FileStoragePath != "" {
		go mem.storeDataWithStoreInterval(dataPathDisk)
	}
	if !config.Restore {
		if _, err := os.Stat(config.FileStoragePath); err == nil {
			os.Remove(config.FileStoragePath)
		}
	}
	return mem
}

func (mem *memStorage) GetCounter(metricName string) (metricValue int64, err error) {
	value, ok := mem.Counter[metricName]
	if ok {
		return value, nil
	}
	return -1, ErrMetricNotFound
}

func (mem *memStorage) GetGauge(metricName string) (metricValue float64, err error) {
	value, ok := mem.Gauge[metricName]
	if ok {
		return value, nil
	}
	return -1, ErrMetricNotFound
}

func (mem *memStorage) GetAll() (map[string]float64, map[string]int64) {
	return mem.Gauge, mem.Counter
}

func (mem *memStorage) UpdateGauge(metricName string, metricValue float64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateGauge func %s", rec)
		}
	}()
	mem.Gauge[metricName] = metricValue
	if mem.config.StoreInterval == 0 && mem.config.DatabaseDSN != "" {
		err := mem.writeToDB()
		if err != nil {
			return err
		}
	} else if mem.config.StoreInterval == 0 && mem.config.FileStoragePath != "" {
		err := mem.writeToDisk()
		if err != nil {
			return err
		}
	}
	return nil
}

func (mem *memStorage) UpdateCounter(metricName string, metricValue int64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateCounter func %s", rec)
		}
	}()
	mem.Counter[metricName] += metricValue
	if mem.config.StoreInterval == 0 && mem.config.DatabaseDSN != "" {
		err := mem.writeToDB()
		if err != nil {
			return err
		}
	} else if mem.config.StoreInterval == 0 && mem.config.FileStoragePath != "" {
		err := mem.writeToDisk()
		if err != nil {
			return err
		}
	}
	return nil
}

func (mem *memStorage) Updates(metrics []utils.Metrics) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			mem.logger.Error("", zap.Any("rec", rec))
			err = fmt.Errorf("error in storagelayer/Updates func %s", rec)
		}
	}()
	for _, metric := range metrics {
		if metric.MType == "gauge" {
			mem.Gauge[metric.ID] = metric.Value
		}
		if metric.MType == "counter" {
			mem.Counter[metric.ID] += metric.Delta
		}
	}
	if mem.config.StoreInterval == 0 && mem.config.DatabaseDSN != "" {
		err := mem.updatesWriteToDB(metrics)
		if err != nil {
			return err
		}
	} else if mem.config.StoreInterval == 0 && mem.config.FileStoragePath != "" {
		err := mem.writeToDisk()
		if err != nil {
			return err
		}
	}
	return nil
}

func (mem *memStorage) storeDataWithStoreInterval(dataPath string) error {
	count := 0
	for {
		if count%mem.config.StoreInterval == 0 {
			if dataPath == dataPathDisk {
				err := mem.writeToDisk()
				if err != nil {
					return err
				}
			}
			if dataPath == dataPathDatabase {
				err := mem.writeToDB()
				if err != nil {
					return err
				}
			}
		}
		count++
		time.Sleep(1 * time.Second)
	}
}

func (mem *memStorage) writeToDisk() error {
	file, err := os.OpenFile(mem.config.FileStoragePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		mem.logger.Fatal("writeToDisk can't open file to store data", zap.Error(err))
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.Encode(mem)
	return nil
}

func (mem *memStorage) readFromDisk() error {
	err := json.Unmarshal([]byte(utils.GetLastLineWithSeek(mem.config.FileStoragePath)), mem)
	if err != nil {
		mem.logger.Error("readFromDisk unmarshal err", zap.Error(err))
		return err
	}
	return nil
}

func (mem *memStorage) retryExecContext(err error, execContext func(ctx context.Context, query string, arg ...any) (sql.Result, error), query string, args ...any) error {
	var pgErr *pgconn.PgError // IsConnectionException //IsInvalidCatalogName
	if errors.As(err, &pgErr) && pgerrcode.IsInvalidCatalogName(pgErr.Code) {
		count := 0
		for {
			time.Sleep(time.Second)
			count++
			if count == 1 || count == 4 || count == 9 {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				_, err = execContext(ctx, query, args)
				if err != nil {
					mem.logger.Error("", zap.Error(err))
					cancel()
					if errors.As(err, &pgErr) && pgerrcode.IsInvalidCatalogName(pgErr.Code) {
						if count == 9 {
							return err
						}
						continue
					}
				} else {
					cancel()
					return nil
				}
			}
			if count == 9 {
				return err
			}
		}
	}
	return nil
}

func (mem *memStorage) CheckDBConnection() error {
	db, err := sql.Open("pgx", mem.config.DatabaseDSN)
	if err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		mem.logger.Error("", zap.Error(err))
		var pgErr *pgconn.PgError // IsConnectionException
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			count := 0
			for {
				time.Sleep(time.Second)
				count++
				if count == 1 || count == 4 || count == 9 {
					ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
					err = db.PingContext(ctx)
					if err != nil {
						cancel()
						if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
							if count == 9 {
								break
							}
							continue
						}
					} else {
						cancel()
						return nil
					}
				}
				if count == 9 {
					break
				}
			}
		}
		return err
	}
	return nil
}

// to do add migrations
func (mem *memStorage) createDBScheme() error {
	db, err := sql.Open("pgx", mem.config.DatabaseDSN)
	if err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	createTableQuery := `CREATE TABLE IF NOT EXISTS metrics (
					   name character(50),
					   type character(25),
					   value double precision,
    				   time_unix integer)`
	_, err = db.ExecContext(ctx, createTableQuery)

	if err != nil {
		errAfterRetry := mem.retryExecContext(err, db.ExecContext, createTableQuery, "")
		if errAfterRetry != nil {
			mem.logger.Fatal("can't create db scheme", zap.Error(err))
		}
	}

	return nil
}

func (mem *memStorage) writeToDB() error {
	db, err := sql.Open("pgx", mem.config.DatabaseDSN)
	if err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	QueryTime := time.Now().Unix()
	insertQuery := `insert into metrics ("name", "type", "value", "time_unix") values ($1, $2, $3, $4);`
	for metricName, metricValue := range mem.Gauge {
		_, err = db.ExecContext(ctx, insertQuery, metricName, "gauge", metricValue, QueryTime)
		if err != nil {
			errAfterRetry := mem.retryExecContext(err, db.ExecContext, insertQuery, metricName, "gauge", metricValue, QueryTime)
			if errAfterRetry != nil {
				mem.logger.Error("", zap.Error(err))
			}
		}
	}
	for metricName, metricValue := range mem.Counter {
		_, err = db.ExecContext(ctx, insertQuery, metricName, "counter", metricValue, QueryTime)
		if err != nil {
			errAfterRetry := mem.retryExecContext(err, db.ExecContext, insertQuery, metricName, "gauge", metricValue, QueryTime)
			if errAfterRetry != nil {
				mem.logger.Error("", zap.Error(err))
			}
		}
	}
	return nil
}

type metrics struct {
	name     string
	format   string
	value    float64
	timeUnix int
}

func (mem *memStorage) readFromDB() error {
	db, err := sql.Open("pgx", mem.config.DatabaseDSN)
	if err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	selectQuery := `select m.name, m.type, m.value, m.time_unix from(select name, max(time_unix)
					as mx from metrics group by name) t join metrics m on m.name = t.name and t.mx = m.time_unix;`
	rows, err := db.QueryContext(ctx, selectQuery)

	if err != nil {
		var pgErr *pgconn.PgError // IsConnectionException
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			count := 0
			for {
				time.Sleep(time.Second)
				count++
				if count == 1 || count == 4 || count == 9 {
					ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
					rows, err = db.QueryContext(ctx, selectQuery)
					if err != nil {
						cancel()
						if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
							if count == 9 {
								return err
							}
							continue
						}
					}
					cancel()
					for rows.Next() {
						var m metrics
						err = rows.Scan(&m.name, &m.format, &m.value, &m.timeUnix)
						if err != nil {
							mem.logger.Error("", zap.Error(err))
							return err
						}
						m.name = strings.Trim(m.name, " ")
						m.format = strings.Trim(m.format, " ")
						if m.format == "gauge" {
							mem.Gauge[m.name] = m.value
						}
						if m.format == "counter" {
							mem.Counter[m.name] = int64(m.value)
						}
					}
					err = rows.Err()
					if err != nil {
						rows.Close()
						mem.logger.Error("", zap.Error(err))
						return err
					}
					rows.Close()
					return nil
				}
				if count == 9 {
					return err
				}
			}
		}
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var m metrics
		err = rows.Scan(&m.name, &m.format, &m.value, &m.timeUnix)
		if err != nil {
			mem.logger.Error("", zap.Error(err))
			return err
		}
		m.name = strings.Trim(m.name, " ")
		m.format = strings.Trim(m.format, " ")
		if m.format == "gauge" {
			mem.Gauge[m.name] = m.value
		}
		if m.format == "counter" {
			mem.Counter[m.name] = int64(m.value)
		}
	}
	err = rows.Err()
	if err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	return nil
}

func (mem *memStorage) updatesWriteToDB(metrics []utils.Metrics) error {
	db, err := sql.Open("pgx", mem.config.DatabaseDSN)
	if err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	tx, err := db.Begin()
	if err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	timeQuery := time.Now().Unix()
	insertQuery := `insert into metrics ("name", "type", "value", "time_unix") values ($1, $2, $3, $4);`
	for _, metric := range metrics {
		if metric.MType == "gauge" {
			_, err := tx.ExecContext(ctx, insertQuery, metric.ID, metric.MType, metric.Value, timeQuery)
			if err != nil {
				errAfterRetry := mem.retryExecContext(err, tx.ExecContext, insertQuery, metric.ID, metric.MType, metric.Value, timeQuery)
				if errAfterRetry != nil {
					mem.logger.Fatal("", zap.Error(errAfterRetry))
					tx.Rollback()
					return errAfterRetry
				}
			}
		}
		if metric.MType == "counter" {
			_, err := tx.ExecContext(ctx, insertQuery, metric.ID, metric.MType, metric.Delta, timeQuery)
			if err != nil {
				errAfterRetry := mem.retryExecContext(err, tx.ExecContext, insertQuery, metric.ID, metric.MType, metric.Value, timeQuery)
				if errAfterRetry != nil {
					mem.logger.Fatal("", zap.Error(errAfterRetry))
					tx.Rollback()
					return errAfterRetry
				}
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		mem.logger.Error("", zap.Error(err))
		return err
	}
	return nil
}
