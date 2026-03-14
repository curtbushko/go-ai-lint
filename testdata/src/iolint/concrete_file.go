// Package iolint contains test cases for the iolint analyzer.
// These test cases detect when concrete types are used instead of io.Reader/io.Writer.
package iolint

import (
	"io"
	"os"
)

// --- BAD: Using concrete *os.File when io.Reader/io.Writer would work ---

// BadConcreteFileReader uses *os.File parameter when io.Reader would suffice.
// This limits the function to only work with files, not buffers, network, etc.
func BadConcreteFileReader(f *os.File) ([]byte, error) { // want "AIL140: parameter uses concrete \\*os.File when io.Reader/io.Writer would suffice"
	return io.ReadAll(f)
}

// BadConcreteFileWriter uses *os.File parameter when io.Writer would suffice.
func BadConcreteFileWriter(f *os.File, data []byte) error { // want "AIL140: parameter uses concrete \\*os.File when io.Reader/io.Writer would suffice"
	_, err := f.Write(data)
	return err
}

// BadConcreteFileReadWriter uses *os.File when only read/write are used.
func BadConcreteFileReadWriter(f *os.File) error { // want "AIL140: parameter uses concrete \\*os.File when io.Reader/io.Writer would suffice"
	buf := make([]byte, 1024)
	_, err := f.Read(buf)
	if err != nil {
		return err
	}
	_, err = f.Write(buf)
	return err
}

// --- GOOD: Using io.Reader/io.Writer interfaces ---

// GoodReaderParam accepts io.Reader, maximizing composability.
func GoodReaderParam(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// GoodWriterParam accepts io.Writer, maximizing composability.
func GoodWriterParam(w io.Writer, data []byte) error {
	_, err := w.Write(data)
	return err
}

// GoodReadWriterParam accepts io.ReadWriter for bidirectional streams.
func GoodReadWriterParam(rw io.ReadWriter) error {
	buf := make([]byte, 1024)
	_, err := rw.Read(buf)
	if err != nil {
		return err
	}
	_, err = rw.Write(buf)
	return err
}

// --- EDGE CASE: Using File-specific methods justifies *os.File ---

// GoodFileWithStat uses Stat() which is specific to *os.File.
// Using *os.File is justified here.
func GoodFileWithStat(f *os.File) (int64, error) {
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GoodFileWithSeek uses Seek() - while io.Seeker exists, this is common.
func GoodFileWithSeek(f *os.File) error {
	_, err := f.Seek(0, io.SeekStart)
	return err
}

// GoodFileWithName uses Name() which is specific to *os.File.
func GoodFileWithName(f *os.File) string {
	return f.Name()
}

// GoodFileWithFd uses Fd() which is specific to *os.File.
func GoodFileWithFd(f *os.File) uintptr {
	return f.Fd()
}

// GoodFileWithSync uses Sync() which is specific to *os.File.
func GoodFileWithSync(f *os.File) error {
	return f.Sync()
}

// GoodFileWithTruncate uses Truncate() which is specific to *os.File.
func GoodFileWithTruncate(f *os.File, size int64) error {
	return f.Truncate(size)
}

// GoodFileWithChmod uses Chmod() which is specific to *os.File.
func GoodFileWithChmod(f *os.File, mode os.FileMode) error {
	return f.Chmod(mode)
}

// GoodFileWithChown uses Chown() which is specific to *os.File.
func GoodFileWithChown(f *os.File, uid, gid int) error {
	return f.Chown(uid, gid)
}

// GoodFileWithReadDir uses ReadDir() which is specific to *os.File for directories.
func GoodFileWithReadDir(f *os.File) ([]os.DirEntry, error) {
	return f.ReadDir(-1)
}
