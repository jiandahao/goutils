package files

import (
	"bufio"
	"io/ioutil"
	"os"
)

// Filter a file filter by filename
type Filter func(string) bool

// GetAllFiles get all files under path dirPath
func GetAllFiles(dirPath string, filter Filter) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)

	for _, fi := range dir {
		if fi.IsDir() {
			dirs = append(dirs, dirPath+PthSep+fi.Name())
			GetAllFiles(dirPath+PthSep+fi.Name(), filter)
		} else {
			if filter != nil {
				if ok := filter(fi.Name()); ok {
					files = append(files, dirPath+PthSep+fi.Name())
				}
			} else {
				files = append(files, dirPath+PthSep+fi.Name())
			}
		}
	}

	for _, table := range dirs {
		temp, _ := GetAllFiles(table, filter)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return files, nil
}

// IsExist returns a boolean indicating whether the error is known to report
// that a file or directory already exists. It is satisfied by ErrExist as well as some syscall errors.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// ReadLines returns a channel that could received file content by lines
func ReadLines(filePath string) (<-chan string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	lines := make(chan string)
	go func() {
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			text := scanner.Text()
			lines <- text
		}

		file.Close()
		close(lines)
	}()

	return lines, nil
}
