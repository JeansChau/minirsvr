package webfs

import (
	"os"

	// "path"

	"fmt"
	"io"
	"io/fs"

	//"log"

	"math"
)

// go mod init voxwebsvr.com/webfs

func calcCeilPowerOfTwo(value float64) float64 {
	return math.Pow(2, math.Ceil(math.Log(value)/math.Ln2))
}
func calcInBufIndex(po2 float64) float64 {
	return math.Log(po2) / math.Ln2
}

var emptyBuf []byte
var bytesContentType = "application/octet-stream"

// 每次 for 循环最多读取 128 kb bytes
var maxSegBufSize float64 = 1024 * 128

var maxBytesSize int = 1024 * 1024 * 128

var bufFlags [32]int
var in_bufs [32]*[]byte
var out_bufs [32]*[]byte

type BytesPtr *[]byte

type BufList struct {
	index int32
	/*
	* buf 的数量
	 */
	size int32
	bufs [128]BytesPtr
	// 当前bug的字节大小
	bufBytesSize int32
	// 实际创建的 buf(bytes 数组) 数量
	bufsTotal int32
}

func (list *BufList) GetSize(buf BytesPtr) int32 {
	return list.size
}
func (list *BufList) PushBuf(buf BytesPtr) {
	if buf != nil {
		list.size++
		// fmt.Println("BufList::PushBuf(): ", list.index, list.bufBytesSize, ", size: ", list.size)
	}
}
func (list *BufList) PopBuf() (BytesPtr, int32) {

	var bufPtr BytesPtr = nil
	var i int32 = -1
	if list.size > 0 {
		list.size--
		i = list.size
		bufPtr = list.bufs[i]
		if bufPtr == nil {
			buf := make([]byte, list.bufBytesSize)
			bufPtr = &buf
			list.bufs[i] = bufPtr
			list.bufsTotal++
			fmt.Println("BufList::PopBuf(): ", list.index, list.bufBytesSize, ",i =", i, ", >>> new, bufsTotal: ", list.bufsTotal)

		} else {
			// fmt.Println("BufList::PopBuf(): ", list.index, list.bufBytesSize, ",i =", i)
		}
	} else {
		fmt.Println("Error, BufList::PopBuf(), list.size == 0.")
	}
	return bufPtr, i
}

func NewBufListPtr(index int32, bufBytesSize int32) *BufList {
	list := BufList{}
	list.index = index
	list.size = 128
	list.bufBytesSize = bufBytesSize
	list.bufsTotal = 0
	return &list
}

type BufPool struct {
	size int32
	list [32]*BufList
}

func (pool *BufPool) GetList(index int32) *BufList {
	listPtr := pool.list[index]
	if listPtr != nil {
		return listPtr
	}
	bytesSize := int32(math.Pow(2, float64(index)))
	fmt.Println("GetList( index =", index, ", bytesSize =", bytesSize, " )")
	listPtr = NewBufListPtr(index, bytesSize)
	pool.list[index] = listPtr
	return listPtr
}
func NewBufPool() BufPool {
	pool := BufPool{}
	pool.size = 32
	return pool
}

var inPool BufPool = NewBufPool()
var outPool BufPool = NewBufPool()

func ReadFileBySteps(pathStr *string, beginPos int, endPos int) *SendBufInfo {

	// fmt.Println("ReadFileBySteps(), pathStr", *pathStr)

	if beginPos < 0 {
		beginPos = 0
	}
	bytesTotalSize := endPos - beginPos

	fmt.Println("ReadFileBySteps(), XXXX beginPos: ", beginPos, ", endPos: ", endPos, ",bytesTotalSize: ", bytesTotalSize)
	sendBuf := emptyBuf
	sendSize := 0
	var fiPtr *fs.FileInfo = nil
	success := true
	if bytesTotalSize > 0 {
		// file, err := os.Open((*pathStr))
		file, err := os.OpenFile((*pathStr), os.O_RDONLY, os.ModeDevice)
		if err == nil {

			fi, _ := file.Stat()
			// fiPtr = &fi
			fileBytesTotal := int(fi.Size())
			if endPos > fileBytesTotal {
				endPos = fileBytesTotal
				bytesTotalSize = endPos - beginPos
			}
			defer file.Close()

			if bytesTotalSize < 0 {
				si := SendBufInfo{&emptyBuf, sendSize, fiPtr, bytesContentType, false, false}
				return &si
			}
			// fmt.Println("ReadFileBySteps(), beginPos:", beginPos, ", endPos", endPos, ",fileBytesTotal: ", fileBytesTotal, ",bytesTotalSize: ", bytesTotalSize)

			bufSizef64 := calcCeilPowerOfTwo(float64(bytesTotalSize))
			segBufSize := bufSizef64
			if segBufSize > maxSegBufSize {
				segBufSize = maxSegBufSize
			}
			readInBufIndex := int32(calcInBufIndex(segBufSize))
			var bufIntSize int64 = int64(segBufSize)
			/*
				readBufPtr := in_bufs[readInBufIndex]

				if readBufPtr == nil {
					tempBuf := make([]byte, bufIntSize)
					readBufPtr = &tempBuf
					in_bufs[readInBufIndex] = readBufPtr
					// fmt.Println("read in -> create new buf(", bufIntSize, "bytes)")
				}
				//*/
			inBufListPtr := inPool.GetList(readInBufIndex)
			inBufPtr, inBuf_index := inBufListPtr.PopBuf()

			bufFlags[readInBufIndex] += 1
			if bufFlags[readInBufIndex] > 1 {
				fmt.Println("ReadFileBySteps(), bufFlags[", readInBufIndex, "]: ", bufFlags[readInBufIndex])
			}
			segSize := bufIntSize
			readInBuf := *inBufPtr

			pos := int64(beginPos)

			if int32(bytesTotalSize) <= inBufListPtr.bufBytesSize {
				count, _ := file.ReadAt(readInBuf, pos)
				sendBuf = readInBuf[:bytesTotalSize]
				sendSize = bytesTotalSize
				// fmt.Println("ReadFileBySteps(), beginPos:", beginPos, ", endPos", endPos, ",sendSize: ", sendSize, ", len(sendBuf): ", len(sendBuf))
				bufFlags[readInBufIndex] -= 1
				fmt.Println("read bytes buf by only one step, pos:", pos, ", inBuf_index:", inBuf_index, ",listIndex:", readInBufIndex, " count: ", count, ",bytesTotalSize:", bytesTotalSize)
			} else {
				fmt.Println("read bytes buf by multi-steps")
				// 存放读取结果
				bufIntSize = int64(bufSizef64)
				readInBufIndex = int32(calcInBufIndex(bufSizef64))
				/*
					outBufPtr := out_bufs[readInBufIndex]
					if outBufPtr == nil {
						tempBuf := make([]byte, bufIntSize)
						outBufPtr = &tempBuf
						out_bufs[readInBufIndex] = outBufPtr
						// fmt.Println("write out -> create new  buf(", bufIntSize, "bytes)")
					}
					//*/

				outBufListPtr := outPool.GetList(readInBufIndex)
				outBufPtr, _ := outBufListPtr.PopBuf()
				writeOutBuf := *outBufPtr

				rbytesSize := 0
				var outPos int = 0
				var segSizeInt = int(segSize)
				for {
					count, err := file.ReadAt(readInBuf, pos)
					size := count + rbytesSize
					if size > bytesTotalSize {
						count -= (size - bytesTotalSize)
					}
					dst := writeOutBuf[outPos : outPos+count]
					src := readInBuf[:count]
					copy(dst, src)
					if err == io.EOF {
						fmt.Println("read in EOF error.")
						break
					}
					rbytesSize += count
					if rbytesSize >= bytesTotalSize {
						break
					}
					pos += segSize
					outPos += segSizeInt
				}
				sendBuf = writeOutBuf[:bytesTotalSize]
				sendSize = bytesTotalSize
				// fmt.Println("ReadFileBySteps(), beginPos:", beginPos, ", endPos", endPos, ",sendSize: ", sendSize, ", len(sendBuf): ", len(sendBuf))
				bufFlags[readInBufIndex] -= 1

				outBufListPtr.PushBuf(outBufPtr)
			}
			inBufListPtr.PushBuf(inBufPtr)
		} else {
			fmt.Println("read in file error.")
			success = false
		}
	} else {
		fmt.Println("bytes range error.")
		success = false
	}
	// return &sendBuf, sendSize, fiPtr, &bytesContentType
	si := SendBufInfo{&sendBuf, sendSize, fiPtr, bytesContentType, false, success}
	return &si
}
