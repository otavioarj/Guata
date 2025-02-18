package main

import (
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func myPrint(str string) {

	glob.outtext.Append(time.Now().Format("15:04:05") + ": " + str + "\n")
}

// Helper function to get Content-Length from the HTTP headers
func getContentLength(headers []byte) int {
	// Define the "Content-Length:" header prefix
	contentLengthPrefix := []byte("Content-Length:")
	// Iterate through headers manually
	lineStart := 0
	for i := 0; i < len(headers); i++ {
		// Look for end of a line "\r\n"
		if headers[i] == '\r' && i+1 < len(headers) && headers[i+1] == '\n' {
			line := headers[lineStart:i]
			lineStart = i + 2 // Skip "\r\n"

			// Compare if the line starts with "Content-Length:"
			if len(line) > len(contentLengthPrefix) && cmpBytes(line[:len(contentLengthPrefix)], contentLengthPrefix) {
				// Skip the "Content-Length:" part and any spaces
				startIdx := len(contentLengthPrefix)
				for startIdx < len(line) && line[startIdx] == ' ' {
					startIdx++ // Skip spaces
				}
				// Now parse the number (digits)
				// not using strconv or bytes to avoid whole packages just to cmp bytes....
				var length int
				for ; startIdx < len(line) && line[startIdx] >= '0' && line[startIdx] <= '9'; startIdx++ {
					length = length*10 + int(line[startIdx]-'0') // Build the number manually
				}
				return length
			}
		}
	}
	// Return 0 if Content-Length is not found
	return 0
}

// Yeah comparing bytes by not using a whole package just for it
func cmpBytes(slice1, slice2 []byte) bool {
	//fmt.Printf("%s %v", slice1, slice1)
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

func monitorResize(isDead *bool, w *fyne.Window, text1, text2 *widget.Entry) {
	var newSize fyne.Size
	var factorH float32
	for {
		if *isDead {
			return // killing this thread :)
		}
		if !text2.Hidden {
			newSize = (*w).Canvas().Size()
			factorH = (newSize.Height/800)*(newSize.Height/800)/10 + 0.66
			text1.Resize(fyne.NewSize(newSize.Width*0.498, newSize.Height*factorH))
			text2.Resize(fyne.NewSize(newSize.Width*0.498, newSize.Height*factorH))
		}
		time.Sleep(450 * time.Millisecond)
	}
}

func sendFunc(sync *insync, hostEntry, requestText, responseText *widget.Entry, sendButton *widget.Button) func() {
	return func() {
		var start int64
		if sendButton.Text == "Stop" {
			sync.syncAbort = true
			sendButton.Text = "Send"
			hostEntry.Enable()
			requestText.Enable()
			myPrint("Tab" + strconv.Itoa(sync.id) + "interrupted")
			return
		}
		sendButton.SetText("Stop")
		hostEntry.Disable()
		responseText.SetText("")
		requestText.Disable()
		requestChan := make(chan []byte)
		var merr string
		if hostEntry.Text == "" {
			if requestText.Text == "" {
				sendButton.SetText("Send")
				sync.syncAbort = false
				hostEntry.Enable()
				requestText.Enable()
				return
			} else if _, ct1, ok := strings.Cut(requestText.Text, "Host: "); ok {
				var result strings.Builder
				for c := 0; c < len(ct1); c++ {
					if ct1[c] == '\n' || ct1[c] == '\r' || ct1[c] == ' ' {
						continue
					}
					result.WriteByte(ct1[c])
				}
				hostEntry.SetText(result.String() + ":443")
			}
		}
		size := len(requestText.Text)
		if size > 0 && requestText.Text[size-4:] != "\r\n\r\n" && requestText.Text[size-2:] != "\n\n" {
			requestText.Append("\r\n\r\n")
		}
		start = time.Now().UnixMilli()
		go sendRequest(hostEntry.Text, requestText.Text, requestChan, &merr, sync)
		go func() {
			for data := range requestChan {
				responseText.Append(string(data))
			}
			if merr != "" {
				if len(merr) > 4 && merr[:3] == "err" {
					responseText.Append("\nError: " + merr[4:])
				}
			}
			if sync.tabs.isJitter {
				myPrint("Tab" + strconv.Itoa(sync.id) + " " + strconv.Itoa(int(time.Now().UnixMilli()-start)) + "ms")
			}
			sendButton.SetText("Send")
			sync.syncAbort = false
			hostEntry.Enable()
			requestText.Enable()
		}()
	}
}
