package filex

import (
	"io"
	"os"
)

func IsBinary(file *os.File) (bool, error) {
	buf := make([]byte, 64)
	_, err := io.ReadFull(file, buf)
	switch err {
	case nil, io.ErrUnexpectedEOF, io.EOF:
	default:
		return false, err
	}
	// 复原文件指针至开头
	file.Seek(0, io.SeekStart)

	for _, v := range buf {
		if v >= 0x20 && v <= 0xff {
			continue
		}
		if v == '\r' || v == '\n' || v == '\t' {
			continue
		}

		if v == '\a' || v == 0 {
			continue
		}
		return true, nil
	}
	return false, nil
}
