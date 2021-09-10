package util

import (
	"crypto/md5"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/common"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func CreateDir(path string) error {
	if IsExist(path) == false {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func CountDirFiles(dirName string) int {
	if !IsDir(dirName) {
		return 0
	}
	var count int
	err := filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		count++
		return nil
	})
	if err != nil {
		logger.Log.Fatalln("count dir files failed %v", err)
		return 0
	}
	return count
}

//FileMD5 count file md5
func FileMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	m := md5.New()
	if _, err := io.Copy(m, file); err != nil {
		return "", err
	}

	fileMd5 := fmt.Sprintf("%x", m.Sum(nil))
	return fileMd5, nil
}

func LocalMd5Sum(src string) string {
	md5Str, err := FileMD5(src)
	if err != nil {
		logger.Log.Fatalf("get file md5 failed %v", err)
		return ""
	}
	return md5Str
}

// MkFileFullPathDir is used to file create the full path.
// eg. there is a file "./aa/bb/xxx.txt", and dir ./aa/bb is not exist, and will create the full path dir.
func MkFileFullPathDir(fileName string) error {
	localDir := filepath.Dir(fileName)
	err := Mkdir(localDir)
	if err != nil {
		return fmt.Errorf("create local dir %s failed: %v", localDir, err)
	}
	return nil
}

func Mkdir(dirName string) error {
	return os.MkdirAll(dirName, os.ModePerm)
}

func WriteFile(fileName string, content []byte) error {
	dir := filepath.Dir(fileName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, common.FileMode0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(fileName, content, common.FileMode0644); err != nil {
		return err
	}
	return nil
}
