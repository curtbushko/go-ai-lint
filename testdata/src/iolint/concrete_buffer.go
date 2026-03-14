// Package iolint contains test cases for the iolint analyzer.
// These test cases detect when concrete *bytes.Buffer is used instead of io.Reader/io.Writer.
package iolint

import (
	"bytes"
	"io"
)

// --- BAD: Using concrete *bytes.Buffer when io.Reader/io.Writer would work ---

// BadConcreteBufferWriter uses *bytes.Buffer parameter when io.Writer would suffice.
// This limits the function to only work with buffers, not files, network, etc.
func BadConcreteBufferWriter(buf *bytes.Buffer, data []byte) error { // want "AIL141: parameter uses concrete \\*bytes.Buffer when io.Reader/io.Writer would suffice"
	_, err := buf.Write(data)
	return err
}

// BadConcreteBufferReader uses *bytes.Buffer parameter when io.Reader would suffice.
func BadConcreteBufferReader(buf *bytes.Buffer) ([]byte, error) { // want "AIL141: parameter uses concrete \\*bytes.Buffer when io.Reader/io.Writer would suffice"
	return io.ReadAll(buf)
}

// BadConcreteBufferReadWriter uses *bytes.Buffer when only Read/Write are used.
func BadConcreteBufferReadWriter(buf *bytes.Buffer) error { // want "AIL141: parameter uses concrete \\*bytes.Buffer when io.Reader/io.Writer would suffice"
	data := make([]byte, 1024)
	_, err := buf.Read(data)
	if err != nil {
		return err
	}
	_, err = buf.Write(data)
	return err
}

// BadConcreteBufferWriteTo uses *bytes.Buffer with WriteTo (implements io.WriterTo).
func BadConcreteBufferWriteTo(buf *bytes.Buffer, w io.Writer) error { // want "AIL141: parameter uses concrete \\*bytes.Buffer when io.Reader/io.Writer would suffice"
	_, err := buf.WriteTo(w)
	return err
}

// BadConcreteBufferReadFrom uses *bytes.Buffer with ReadFrom (implements io.ReaderFrom).
func BadConcreteBufferReadFrom(buf *bytes.Buffer, r io.Reader) error { // want "AIL141: parameter uses concrete \\*bytes.Buffer when io.Reader/io.Writer would suffice"
	_, err := buf.ReadFrom(r)
	return err
}

// BadMultipleBufferParams has multiple *bytes.Buffer params, all using only Write.
func BadMultipleBufferParams(buf1 *bytes.Buffer, buf2 *bytes.Buffer) error { // want "AIL141: parameter uses concrete \\*bytes.Buffer when io.Reader/io.Writer would suffice" "AIL141: parameter uses concrete \\*bytes.Buffer when io.Reader/io.Writer would suffice"
	_, err := buf1.Write([]byte("hello"))
	if err != nil {
		return err
	}
	_, err = buf2.Write([]byte("world"))
	return err
}

// --- GOOD: Using io.Reader/io.Writer interfaces ---

// GoodWriterParam accepts io.Writer, maximizing composability.
func GoodBufferWriterParam(w io.Writer, data []byte) error {
	_, err := w.Write(data)
	return err
}

// GoodReaderParam accepts io.Reader, maximizing composability.
func GoodBufferReaderParam(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// --- EDGE CASE: Using Buffer-specific methods justifies *bytes.Buffer ---

// GoodBufferWithBytes uses Bytes() which is specific to *bytes.Buffer.
// Using *bytes.Buffer is justified here.
func GoodBufferWithBytes(buf *bytes.Buffer) []byte {
	return buf.Bytes()
}

// GoodBufferWithString uses String() which is specific to *bytes.Buffer.
func GoodBufferWithString(buf *bytes.Buffer) string {
	return buf.String()
}

// GoodBufferWithLen uses Len() which is specific to *bytes.Buffer.
func GoodBufferWithLen(buf *bytes.Buffer) int {
	return buf.Len()
}

// GoodBufferWithCap uses Cap() which is specific to *bytes.Buffer.
func GoodBufferWithCap(buf *bytes.Buffer) int {
	return buf.Cap()
}

// GoodBufferWithReset uses Reset() which is specific to *bytes.Buffer.
func GoodBufferWithReset(buf *bytes.Buffer) {
	buf.Reset()
}

// GoodBufferWithGrow uses Grow() which is specific to *bytes.Buffer.
func GoodBufferWithGrow(buf *bytes.Buffer, n int) {
	buf.Grow(n)
}

// GoodBufferWithTruncate uses Truncate() which is specific to *bytes.Buffer.
func GoodBufferWithTruncate(buf *bytes.Buffer, n int) {
	buf.Truncate(n)
}

// GoodBufferWithNext uses Next() which is specific to *bytes.Buffer.
func GoodBufferWithNext(buf *bytes.Buffer, n int) []byte {
	return buf.Next(n)
}

// GoodBufferWithUnreadByte uses UnreadByte() which is specific to *bytes.Buffer.
func GoodBufferWithUnreadByte(buf *bytes.Buffer) error {
	_, _ = buf.ReadByte()
	return buf.UnreadByte()
}

// GoodBufferWithUnreadRune uses UnreadRune() which is specific to *bytes.Buffer.
func GoodBufferWithUnreadRune(buf *bytes.Buffer) error {
	_, _, _ = buf.ReadRune()
	return buf.UnreadRune()
}

// GoodBufferWithReadBytes uses ReadBytes() which is specific to *bytes.Buffer.
func GoodBufferWithReadBytes(buf *bytes.Buffer, delim byte) ([]byte, error) {
	return buf.ReadBytes(delim)
}

// GoodBufferWithReadString uses ReadString() which is specific to *bytes.Buffer.
func GoodBufferWithReadString(buf *bytes.Buffer, delim byte) (string, error) {
	return buf.ReadString(delim)
}

// GoodBufferWithWriteString uses WriteString() - while it's part of io.StringWriter,
// combined with Bytes() justifies *bytes.Buffer.
func GoodBufferWithWriteStringAndBytes(buf *bytes.Buffer, s string) []byte {
	_, _ = buf.WriteString(s)
	return buf.Bytes()
}

// GoodBufferWithWriteByte uses WriteByte().
func GoodBufferWithWriteByte(buf *bytes.Buffer, c byte) error {
	return buf.WriteByte(c)
}

// GoodBufferWithWriteRune uses WriteRune().
func GoodBufferWithWriteRune(buf *bytes.Buffer, r rune) error {
	_, err := buf.WriteRune(r)
	return err
}

// GoodBufferWithReadByte uses ReadByte().
func GoodBufferWithReadByte(buf *bytes.Buffer) (byte, error) {
	return buf.ReadByte()
}

// GoodBufferWithReadRune uses ReadRune().
func GoodBufferWithReadRune(buf *bytes.Buffer) (rune, int, error) {
	return buf.ReadRune()
}

// GoodBufferWithAvailable uses Available() which is specific to *bytes.Buffer.
func GoodBufferWithAvailable(buf *bytes.Buffer) int {
	return buf.Available()
}

// GoodBufferWithAvailableBuffer uses AvailableBuffer() which is specific to *bytes.Buffer.
func GoodBufferWithAvailableBuffer(buf *bytes.Buffer) []byte {
	return buf.AvailableBuffer()
}

// GoodBufferMixed uses Write but also Bytes - justified.
func GoodBufferMixed(buf *bytes.Buffer) []byte {
	_, _ = buf.Write([]byte("hello"))
	return buf.Bytes()
}
