package appUtils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/relepega/doujinstyle-downloader/internal/store"
)

var re = regexp.MustCompile(`\.$`)

func MkdirAll(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			return fmt.Errorf("Error creating folder: %v", err)
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

func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return true
	}
	return false
}

func GetAppTempDir() string {
	v, err := store.GetStore().Get("tempdir")
	if err != nil {
		return "."
	}

	tempdir, _ := v.(string)

	return tempdir
}

func GenerateRandomFilename() string {
	// Generate a random string of 16 characters
	b := make([]byte, 16)
	rand.Read(b)

	// Convert the bytes to a hex string
	filename := fmt.Sprintf("%x", b)

	return filename
}

// https://stackoverflow.com/questions/23558425
func GetLocalIPAddr() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	// if error, return fallback address
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func DownloadFile(
	url,
	tempDir,
	finalFilepath string,
	setProgress func(p int8),
) (err error) {
	if setProgress == nil {
		return fmt.Errorf("DownloadFile: setProgress cannot be nil")
	}

	exists, err := FileExists(finalFilepath)
	if err != nil {
		return err
	}

	if exists {
		setProgress(100)
		return nil
	}

	// write to a temp file first to avoid incomplete downloads
	tempf, err := os.CreateTemp(tempDir, "*")
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
	var totalSize int64
	totalSize, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		totalSize = -1
	}

	// Create a buffer for copying
	buf := make([]byte, 32*1024)

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
		var currentProgress int8
		if totalSize == -1 {
			currentProgress = 127
		} else {
			currentProgress = int8((float64(currentSize) / float64(totalSize)) * 100)
		}

		setProgress(currentProgress)

		// Write the chunk to the temp file
		_, err := tempf.Write(buf[:n])
		if err != nil {
			return err
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
	out, err := os.Create(finalFilepath)
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

func CleanString(s string) string {
	trimmed := strings.TrimSpace(s)
	clean := strings.ReplaceAll(trimmed, "\n", "")

	return clean
}
