package compress

import (
	"compress/gzip"
	"io"
)

// Compress compress data from given reader and write into given writer
func Compress(from io.Reader, to io.Writer) error {
	writer := gzip.NewWriter(to)
	defer writer.Close() // 一定要Close

	buf := make([]byte, 1024)
	for {
		count, err := from.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		currentBytes := buf[:count]
		if _, err := writer.Write(currentBytes); err != nil {
			return err
		}
		writer.Flush() // 手动flush，否则得不到数据
	}

	return nil
}

// Decompress decompress data from given reader and then write into given writer
func Decompress(from io.Reader, to io.Writer) error {
	reader, err := gzip.NewReader(from)
	if err != nil {
		return err
	}
	defer reader.Close()

	buf := make([]byte, 1024)
	for {
		count, err := reader.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		currentBytes := buf[:count]
		if _, err := to.Write(currentBytes); err != nil {
			return err
		}
	}
}
