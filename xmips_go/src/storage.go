package main

import (
	"os"
	"path/filepath"
)

// Storage 存储器访问类，对应 C++ 的 storage 类
type Storage struct {
	ID       int
	pos      string
	visitWay string
}

func newStorage(key int, store string) *Storage {
	return &Storage{ID: key, pos: store}
}

// getFile 打开文件，choice: 0=read, 1=write, 2=read+
func (s *Storage) getFile(filename string, choice int) (*os.File, int) {
	tmp := filepath.Join(s.pos, filename)
	var mode string
	switch choice {
	case 0:
		mode = "r"
	case 1:
		mode = "w"
	case 2:
		mode = "r+"
	default:
		RES = 80
		printf("D%-5d visit way:%d ", s.ID, choice)
		systemChecker.check(RES, sysLog[:])
		return nil, -1
	}

	var f *os.File
	var err error
	if mode == "w" {
		f, err = os.Create(tmp)
	} else {
		f, err = os.Open(tmp)
	}
	if err != nil {
		RES = 81
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d file path:\"%s\" ", s.ID, tmp)
			systemChecker.check(RES, sysLog[:])
		}
		return nil, -1
	}
	s.visitWay = mode
	return f, 0
}

// releaseFile 关闭文件
func (s *Storage) releaseFile(f *os.File) int {
	if f != nil {
		f.Close()
		return 0
	}
	RES = 82
	if systemChecker.showLevel(RES, sysLog[:]) {
		printf("D%-5d ", s.ID)
		systemChecker.check(RES, sysLog[:])
	}
	return -1
}
