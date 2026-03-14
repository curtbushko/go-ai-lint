// Package iolint contains test cases for the iolint analyzer.
// These test cases detect when concrete *http.Response is used instead of io.ReadCloser.
package iolint

import (
	"io"
	"net/http"
)

// --- BAD: Using *http.Response when only Body is accessed ---

// BadConcreteResponseBodyOnly uses *http.Response parameter when io.ReadCloser would suffice.
// This limits the function's testability and composability.
func BadConcreteResponseBodyOnly(resp *http.Response) ([]byte, error) { // want "AIL142: parameter uses concrete \\*http.Response when only Body is accessed; consider using io.ReadCloser"
	return io.ReadAll(resp.Body)
}

// BadConcreteResponseBodyClose uses *http.Response when only Body and Close are used.
func BadConcreteResponseBodyClose(resp *http.Response) error { // want "AIL142: parameter uses concrete \\*http.Response when only Body is accessed; consider using io.ReadCloser"
	defer resp.Body.Close()
	_, err := io.ReadAll(resp.Body)
	return err
}

// BadMultipleResponseParams has multiple *http.Response params, all using only Body.
func BadMultipleResponseParams(resp1 *http.Response, resp2 *http.Response) error { // want "AIL142: parameter uses concrete \\*http.Response when only Body is accessed; consider using io.ReadCloser" "AIL142: parameter uses concrete \\*http.Response when only Body is accessed; consider using io.ReadCloser"
	_, err := io.ReadAll(resp1.Body)
	if err != nil {
		return err
	}
	_, err = io.ReadAll(resp2.Body)
	return err
}

// --- GOOD: Using io.ReadCloser interface ---

// GoodReadCloserParam accepts io.ReadCloser, maximizing composability.
func GoodReadCloserParam(body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	return io.ReadAll(body)
}

// GoodReaderParam accepts io.Reader when Close is not needed.
func GoodResponseReaderParam(body io.Reader) ([]byte, error) {
	return io.ReadAll(body)
}

// --- EDGE CASE: Using Response-specific fields justifies *http.Response ---

// GoodResponseWithStatusCode uses StatusCode which is specific to *http.Response.
// Using *http.Response is justified here.
func GoodResponseWithStatusCode(resp *http.Response) (int, error) {
	defer resp.Body.Close()
	_, err := io.ReadAll(resp.Body)
	return resp.StatusCode, err
}

// GoodResponseWithStatus uses Status which is specific to *http.Response.
func GoodResponseWithStatus(resp *http.Response) string {
	return resp.Status
}

// GoodResponseWithHeader uses Header which is specific to *http.Response.
func GoodResponseWithHeader(resp *http.Response) string {
	return resp.Header.Get("Content-Type")
}

// GoodResponseWithContentLength uses ContentLength which is specific to *http.Response.
func GoodResponseWithContentLength(resp *http.Response) int64 {
	return resp.ContentLength
}

// GoodResponseWithTransferEncoding uses TransferEncoding which is specific to *http.Response.
func GoodResponseWithTransferEncoding(resp *http.Response) []string {
	return resp.TransferEncoding
}

// GoodResponseWithClose uses Close field which is specific to *http.Response.
func GoodResponseWithClose(resp *http.Response) bool {
	return resp.Close
}

// GoodResponseWithUncompressed uses Uncompressed which is specific to *http.Response.
func GoodResponseWithUncompressed(resp *http.Response) bool {
	return resp.Uncompressed
}

// GoodResponseWithTrailer uses Trailer which is specific to *http.Response.
func GoodResponseWithTrailer(resp *http.Response) http.Header {
	return resp.Trailer
}

// GoodResponseWithRequest uses Request which is specific to *http.Response.
func GoodResponseWithRequest(resp *http.Response) *http.Request {
	return resp.Request
}

// GoodResponseWithTLS uses TLS which is specific to *http.Response.
func GoodResponseWithTLS(resp *http.Response) bool {
	return resp.TLS != nil
}

// GoodResponseWithProtoMajor uses ProtoMajor which is specific to *http.Response.
func GoodResponseWithProtoMajor(resp *http.Response) int {
	return resp.ProtoMajor
}

// GoodResponseWithProtoMinor uses ProtoMinor which is specific to *http.Response.
func GoodResponseWithProtoMinor(resp *http.Response) int {
	return resp.ProtoMinor
}

// GoodResponseWithProto uses Proto which is specific to *http.Response.
func GoodResponseWithProto(resp *http.Response) string {
	return resp.Proto
}

// GoodResponseMixed uses Body but also StatusCode - justified.
func GoodResponseMixed(resp *http.Response) ([]byte, int, error) {
	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}
