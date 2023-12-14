package downloader

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/playwright-community/playwright-go"
)

func sanitizePath(s string) string {
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

func getExhibitions(strVal string) string {
	re := regexp.MustCompile("^(C[0-9]+)|(M[0-9]-[0-9]+)|(AC[0-9])$")
	matches := []string{}

	for _, substr := range strings.Split(strVal, ", ") {
		if re.MatchString(substr) {
			matches = append(matches, substr)
		}
	}

	var fullStr string

	if len(matches) == 0 {
		fullStr = ""
	} else {
		fullStr = " [" + strings.Join(matches, ", ") + "]"
	}

	return fullStr
}

func CraftFilename(page playwright.Page) (string, error) {
	album, err := page.Evaluate("document.querySelector('h2').innerText")
	if err != nil {
		return "", err
	}

	artist, err := page.Evaluate("document.querySelectorAll('.pageSpan2')[0].innerText")
	if err != nil {
		return "", err
	}

	format, err := page.Evaluate(`
	   Array.from(document.querySelectorAll(".pageSpan1")).find(el => el.innerText == "Format:").nextElementSibling.innerText
	`)
	if err != nil {
		return "", err
	}

	val, err := page.Evaluate("document.querySelectorAll('.pageSpan2')[1].innerText")
	if err != nil {
		return "", err
	}
	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string: %v", val)
	}
	event := getExhibitions(strVal)

	return sanitizePath(fmt.Sprintf("%s — %s%s [%s]", artist, album, event, format)), nil
}
