package wal

import (
	"encoding/binary"
	"io"
	"os"
)

type FileWAL struct {
	writePath string
	writeFile *os.File
}

// create a new FileWAL instance
func NewFileWAL(writePath string) (*FileWAL, error) {
	file, err := os.OpenFile(writePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &FileWAL{
		writePath: writePath,
		writeFile: file,
	}, nil
}

// write to File
func (fw *FileWAL) WriteOp(opType uint8, key string, value []byte) error {
	// Implement the logic to write the operation to the file
	keyLen := uint32(len(key))
	valueLen := uint32(len(value))
	// 1 opType, 4 bytes for key length, 4 bytes for value length, key bytes, value bytes
	totalLen := 1 + 4 + 4 + keyLen + valueLen // 1 byte for opType, 4 bytes for key length, 4 bytes for value length
	// write the operation type 1 byte
	buf := make([]byte, totalLen)
	buf[0] = opType
	// write key length 4 bytes
	binary.BigEndian.PutUint32(buf[1:5], keyLen)
	// write value length 4 bytes
	binary.BigEndian.PutUint32(buf[5:9], valueLen)
	// write key bytes
	copy(buf[9:9+keyLen], key)
	// write value bytes
	copy(buf[9+keyLen:], value)
	// write to the file
	_, err := fw.writeFile.Write(buf)
	if err != nil {
		return err
	}
	// call sync for durability
	return fw.writeFile.Sync()
}

func (f *FileWAL) ReadWAL() (entries []LogEntry, err error) {
	// Implement the logic to read the WAL file and return the log entries
	file, err := os.Open(f.writePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	// store the header bytes
	headerBuf := make([]byte, 9) // 1 byte for opType, 4 bytes for key length, 4 bytes for value length
	// read the file using full file
	for {
		_, err = io.ReadFull(file, headerBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		// read the header bytes and extract the opType, key length, and value length
		opType := headerBuf[0]
		keyLen := binary.BigEndian.Uint32(headerBuf[1:5])
		valueLen := binary.BigEndian.Uint32(headerBuf[5:9])
		payloadSize := keyLen + valueLen
		payloadBuf := make([]byte, payloadSize)
		_, err = io.ReadFull(file, payloadBuf)
		if err != nil {
			if err == io.EOF {
				// TODO :throw corrupted log entry error
				return nil, nil
			}
			return nil, err
		}
		key := string(payloadBuf[:keyLen])
		value := make([]byte, valueLen)
		copy(value, payloadBuf[keyLen:])
		entry := LogEntry{
			OpType: opType,
			Key:    key,
			Value:  value,
		}
		entries = append(entries, entry)

	}
	return entries, nil
}

func ReadWALFile(path string) ([]LogEntry, error) {
	return (&FileWAL{writePath: path}).ReadWAL()
}

func (f *FileWAL) Close() error {
	if f.writeFile != nil {
		return f.writeFile.Close()
	}
	return nil
}
