package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"net"
	"time"
)

func isTLSHandshakeFailure(err error) bool {
	var tlsErr *tls.RecordHeaderError
	if err == nil {
		return false
	}
	if errors.As(err, &tlsErr) {
		return true
	}
	return err.Error() == "EOF"
}

func sendRequest(host string, request string, reqchan chan []byte, merr *string, sync *insync) {
	var conn net.Conn
	var size int
	tmp := make([]byte, 1024)
	var headers, body []byte
	headerOk := false
	var done chan bool
	tlsConfig := &tls.Config{
		InsecureSkipVerify: sync.tabs.isTLS,
	}
	defer close(reqchan)

	// Establish TLS connection or fall back to plain TCP
	conn, err := tls.Dial("tcp", host, tlsConfig)
	if isTLSHandshakeFailure(err) {
		conn, err = net.Dial("tcp", host)
	}
	if err != nil {
		*merr = "err conn: " + err.Error()
		return
	}
	defer conn.Close()

	// Send the HTTP request
	_, err = conn.Write([]byte(request))
	if err != nil {
		*merr = "err request: " + err.Error()
		return
	}

	var totalBodySize, bodyRead int
	// Thread (goroutine) to monitor user interruption
	go func() {
		for {
			if *sync.synDead || sync.syncAbort {
				*merr = "err user aborted"
				sync.syncAbort = false
				done <- true
			}
			time.Sleep(time.Millisecond * 850)
		}
	}()
	// Read response in chunks
	for {
		// Non-blocking socket read by channel-interruption
		select {
		case <-done:
			return
		default:
			// Read data into the temporary buffer
			size, err = conn.Read(tmp)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					*merr = "-1"
					return
				}
				if err.Error() == "EOF" {
					*merr = "0"
					return
				} else {
					*merr = "err: response" + err.Error()
					return
				}
			}
			data := tmp[:size]
			// If headers haven't been processed yet, find the header-body separator
			if !headerOk {
				// Look for the separator "\r\n\r\n" to separate header and body
				endOfHeader := bytes.Index(data, []byte("\r\n\r\n"))
				if endOfHeader != -1 {
					// If found, separate header and body
					headers = append(headers, data[:endOfHeader+4]...) // Include "\r\n\r\n" in the header
					body = append(body, data[endOfHeader+4:]...)       // Body starts after the header

					// Parse the Content-Length header if available
					contentLength := getContentLength(headers)
					if contentLength > 0 {
						totalBodySize = contentLength
						bodyRead = len(body)
					}

					// Send header once it's fully read and reset the buffer
					reqchan <- append([]byte(nil), headers...)
					headers = nil   // Reset the headers after sending
					headerOk = true // Mark the headers is ok
				} else {
					headers = append(headers, data...) // Accumulate header data
					continue                           // Proceed to the next read iteration
				}
			} else {
				// Increment body read size and accumulate body data
				bodyRead += len(data)
				body = append(body, data...)
			}

			// Send the body incrementally as data arrives
			if len(body) > 0 {
				reqchan <- append([]byte(nil), body...)
				body = nil // Reset the body after sending
			}

			// Check if we've read enough bytes to finish the body (EOF or Content-Length reached)
			if totalBodySize > 0 && bodyRead >= totalBodySize {
				return
			}

			// Check for end of chunked transfer encoding (e.g., "\r\n\r\n")
			if len(body) >= 4 && cmpBytes(body[len(body)-4:], []byte{'\r', '\n', '\r', '\n'}) {
				return
			}
		}
	}
}
