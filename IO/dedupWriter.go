package IO

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"os"
)

type DedupWriter struct {
	OutputFile *os.File
	CurrentOffset int
	batchCounter int
	maxBatch int
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func NewDedupWriter(filePath string, maxChunksInBatch, chunkMaxSize int) (*DedupWriter, error) {
	outputFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	writer := bufio.NewWriter(outputFile)
	buf := &bytes.Buffer{}
	buf.Grow(chunkMaxSize * maxChunksInBatch)
	dedupWriter := DedupWriter{
		outputFile,
		0,
		0,
		maxChunksInBatch,
		buf,
		writer,
	}
	return &dedupWriter, err
}

func (dedupWriter *DedupWriter) Close()  error {
	dedupWriter.OutputFile.Close()
	return nil
}

func (writer *DedupWriter) WriteBlank(data *[]byte) (int, error) {
	return writer.buffer.Write(*data)
}


// WriteData
// return the number of bytes which were written
func (writer *DedupWriter) WriteData(data *[]byte) (int, error) {
	if writer.batchCounter > writer.maxBatch {
		writer.FlushData()
	}
	// calculate length
	length := len(*data)
	bytesToWrite := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytesToWrite, uint32(length))
	bytesToWrite = append(bytesToWrite, *data...)
	writer.batchCounter++
	writer.buffer.Write(bytesToWrite)
	return length+4, nil
}

func (writer *DedupWriter) WriteMataData(offsetsArr []int) (int, error) {
	for _, offset := range offsetsArr {
		if writer.batchCounter > writer.maxBatch {
			writer.FlushData()
		}
		bytesToWrite := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytesToWrite, uint32(offset))
		writer.batchCounter++
		writer.buffer.Write(bytesToWrite)
	}
	return len(offsetsArr), nil
}

func (writer *DedupWriter) WriteMataDataOffset(offset int) (int, error) {
	bytesToWrite := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytesToWrite, uint32(offset))
	writer.OutputFile.Seek(0,0)
	ioWriter :=  bufio.NewWriter(writer.OutputFile)
	ioWriter.Write(bytesToWrite)
	return 4, nil
}


func (dedupWriter *DedupWriter) FlushData() error {
	dedupWriter.writer.Write(dedupWriter.buffer.Bytes()) //TODO handle error
	logrus.Infof("Wrote %d Bytes to compressed file", len(dedupWriter.buffer.Bytes()))
	dedupWriter.buffer.Reset()
	dedupWriter.batchCounter = 0
	return nil
}

func (dedupWriter *DedupWriter) FlushAll() error {
	dedupWriter.FlushData()
	dedupWriter.writer.Flush()
	return nil
}
