package appUtils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
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

func DirectoryExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
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
		n, readErr := resp.Body.Read(buf)
		if readErr != nil && readErr != io.EOF {
			return readErr
		}

		// this is here just in case there's the need for this check
		// if n == 0 {
		// 	continue
		// }

		// Update the current size
		currentSize += int64(n)

		// Calculate and update the progress
		*progress = int8((float64(currentSize) / float64(totalSize)) * 100)

		// Write the chunk to the temp file
		_, err := tempf.Write(buf[:n])
		if err != nil {
			return readErr
		}

		if readErr == io.EOF {
			break
		}
	}

	tempfn := tempf.Name()

	// Check if the total size matches the Content-Length header
	if currentSize != totalSize {
		tempf.Close()
		os.Remove(tempfn)

		return fmt.Errorf("downloaded file size differs from the one reported by the server")
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
	tempf.Close()
	os.Remove(tempfn)

	return nil
}

func SanitizePath(s string) string {
	r := strings.NewReplacer(
		"\\",
		"١",
		"*",
		"＊",
		"/",
		"∕",
		">",
		"˃",
		"<",
		"˂",
		":",
		"˸",
		"|",
		"-",
		"\"",
		"ˮ",
		"?",
		"？",
	)
	sb := strings.Builder{}

	for _, c := range s {
		sb.WriteString(r.Replace(string(c)))
	}

	res := strings.TrimRight(sb.String(), " ")
	res = strings.TrimLeft(res, " ")

	// replace all the dots only at the end of the string
	re := regexp.MustCompile(`\.$`)
	res = re.ReplaceAllString(res, "ˌ")

	return res
}

func ParseJson[T any](url string, data *T) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}

	return nil
}
