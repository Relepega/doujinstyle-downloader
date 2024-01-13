package appUtils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

func CreateFolder(fname string) error {
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		err = os.MkdirAll(fname, 0755)
		if err != nil {
			fmt.Println("Error creating folder:", err)
			return err
		}
	}

	return nil
}

func FileExists(fp string) (bool, error) {
	_, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func DownloadFile(fp string, url string, progress *int8) (err error) {
	// write to a temp file first to avoid incomplete downloads
	tempf, err := os.CreateTemp("", "doujinstyleDownloader-")
	if err != nil {
		return err
	}
	defer tempf.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad HTTP status: %s", resp.Status)
	}

	// Get the total size of the file
	totalSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)

	// Create a buffer for copying
	buf := make([]byte, 1024)

	// Initialize the current size to zero
	var currentSize int64

	// Copy chunk by chunk
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF {
			break
		}

		// Update the current size
		currentSize += int64(n)

		// Calculate and update the progress
		*progress = int8((float64(currentSize) / float64(totalSize)) * 100)

		// Write the chunk to the temp file
		_, err = tempf.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	// Reset the file pointer to the beginning of the file
	_, err = tempf.Seek(0, 0)
	if err != nil {
		return err
	}

	// Copy content to final location
	out, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, tempf)

	// delete temp file
	tempfn := tempf.Name()
	tempf.Close()
	os.Remove(tempfn)

	return nil
}
