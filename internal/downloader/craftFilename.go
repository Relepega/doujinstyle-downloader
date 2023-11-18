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

	var event string

	re := regexp.MustCompile("C[0-9]+|M[0-9]-[0-9]+")
	val, err := page.Evaluate("document.querySelectorAll('.pageSpan2')[1].innerText")
	if err != nil {
		return "", err
	}
	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string: %v", val)
	}
	matches := re.FindAllString(strVal, -1)

	if len(matches) == 0 {
		event = ""
	} else {
		event = " [" + strings.Join(matches, ", ") + "]"
	}

	return sanitizePath(fmt.Sprintf("%s — %s%s [%s]", artist, album, event, format)), nil
}
