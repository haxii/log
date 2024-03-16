package log

import (
	"compress/gzip"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

func OpenLogFile(logDir, logName string) (*os.File, error) {
	if err := createDirectories(logDir); err != nil {
		return nil, err
	}

	logFilePath, err := backupFile(logDir, logName)
	fileFlag := os.O_CREATE | os.O_WRONLY
	if err != nil {
		fileFlag = fileFlag | os.O_APPEND // fail to back up, append it
	} else {
		fileFlag = fileFlag | os.O_TRUNC // backed up, truncate it
	}

	return os.OpenFile(logFilePath, fileFlag, 0644)
}

func createDirectories(dir string) error {
	if fi, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "check directory failed")
		}
		if err = os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrap(err, "create directory failed")
		}
	} else if !fi.IsDir() {
		return errors.New("destination path is not directory")
	}
	return nil
}

func backupFile(logDir, logName string) (logFilePath string, err error) {
	logFileName := logName + ".log"
	logFilePath = filepath.Join(logDir, logFileName)

	f, fErr := os.Open(logFilePath)
	if fErr != nil {
		return //  we don't care this
	}
	defer f.Close()

	backupPath := filepath.Join(logDir, logName+"."+time.Now().Format("2006-01-02-15.04.05.999")+".log.gz")
	g, gErr := os.Create(backupPath)
	if gErr != nil {
		fmt.Printf("fail to create log backup file %s with error %s\n", backupPath, gErr)
		err = gErr
		return
	}
	defer g.Close()

	w := gzip.NewWriter(g)
	defer w.Close()
	if _, err = io.Copy(w, f); err != nil {
		fmt.Printf("fail to backup file %s with error %s\n", backupPath, err)
		return
	}
	return
}
