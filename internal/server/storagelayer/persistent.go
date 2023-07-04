package storagelayer

import (
	"encoding/json"
	"github.com/nktau/monitoring-service/internal/server/config"
	"github.com/nktau/monitoring-service/internal/server/utils"
	"go.uber.org/zap"
	"os"
	"time"
)

func (mem *memStorage) writeToDiskWithStoreInterval() error {
	count := 0
	for {
		if count%config.Config.StoreInterval == 0 {
			err := mem.writeToDisk()
			if err != nil {
				return err
			}
		}
		count++
		time.Sleep(1 * time.Second)
	}
}

func (mem *memStorage) writeToDisk() error {
	file, err := os.OpenFile(config.Config.FileStoragePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
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
	err := json.Unmarshal([]byte(utils.GetLastLineWithSeek(config.Config.FileStoragePath)), mem)
	if err != nil {
		mem.logger.Error("readFromDisk unmarshal err", zap.Error(err))
		return err
	}
	return nil
}
