package services

import "testing"

type getExhibitionsTest struct {
	test, expected string
}

var getExhibitionsTests = []getExhibitionsTest{
	{
		"",
		"",
	},
	{
		"AC2, M3-46, C46, Rock, J-pop, C4-bbbs, C6, C65675b, M46, M2-543, M-323, AC, AC2321b",
		" [AC2, C6, M2-543]",
	},
}

func TestGetExhibition(t *testing.T) {
	d := &doujinstyle{}
	for i, test := range getExhibitionsTests {
		output := d.getExhibitions(test.test)
		if output != test.expected {
			t.Errorf(
				"downloader.getExhibitions [%d/%d] Got \"%v\", expected \"%v\"",
				i+1,
				len(getExhibitionsTests),
				output,
				test.expected,
			)
		}
	}
}
