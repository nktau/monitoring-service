package utils

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"os"
	"strconv"
)

var logger = InitLogger()

func MetricValueWithoutTrailingZero(metricValue float64) string {
	return strconv.FormatFloat(metricValue, 'f', -1, 64)
}

func GetLastLineWithSeek(filepath string) string {
	fileHandle, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		logger.Fatal("can't open file", zap.Error(err))
	}
	defer fileHandle.Close()
	line := ""
	var cursor int64 = 0
	stat, _ := fileHandle.Stat()
	filesize := stat.Size()
	for {
		cursor -= 1
		fileHandle.Seek(cursor, io.SeekEnd)
		char := make([]byte, 1)
		fileHandle.Read(char)
		if filesize == 0 {
			break
		}
		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			break
		}
		line = fmt.Sprintf("%s%s", string(char), line)
		if cursor == -filesize {
			break
		}
	}
	return line
}
