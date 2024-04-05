package osx

import (
	"os"
	"path/filepath"
)

func CreateFileIfNotExists(fPath string) error {

	// 首先判断文件夹存不存在
	fPathDir := filepath.Dir(fPath)
	if exists, err := CheckFileWhetherExist(fPathDir); err != nil {
		return err
	} else if !exists {
		if err := os.MkdirAll(fPathDir, os.ModePerm); err != nil {
			return err
		}
	}

	if exists, err := CheckFileWhetherExist(fPath); err != nil {
		return err
	} else if !exists {
		file, err := os.Create(fPath)
		if err != nil {
			return err
		}
		file.Close()
	}
	return nil
}

// 检查文件是否存在
func CheckFileWhetherExist(fPath string) (bool, error) {
	_, err := os.Stat(fPath)
	if err == nil {
		return true, nil
	}

	if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}
