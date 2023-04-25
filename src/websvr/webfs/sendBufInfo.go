package webfs

import (
	"io/fs"
)

const (
	SendBytes = 1001
)

type SendBufInfo struct {
	BufPtr      *[]byte
	Size        int
	FiPtr       *fs.FileInfo
	ContentType string
	GzipEnabled bool
	Success     bool
}
