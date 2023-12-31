package appUtils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

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

func DownloadFile(fp string, url string) (err error) {
	// write to a temp file first to avoid incomplete downloads
	tempf, err := os.CreateTemp("", "doujinstyleDownloader-")
	if err != nil {
		return err
	}
	defer tempf.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad HTTP status: %s", resp.Status)
	}

	_, err = io.Copy(tempf, resp.Body)
	if err != nil {
		return err
	}

	// reset the file pointer to the beginning of the file
	_, err = tempf.Seek(0, 0)
	if err != nil {
		return err
	}

	// copy content to final location
	out, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer out.Close()

	io.Copy(out, tempf)

	return nil
}
