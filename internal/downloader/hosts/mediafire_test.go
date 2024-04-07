package hosts

import (
	"path/filepath"
	"testing"

	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"
)

func TestMediafireGetFolderKey(t *testing.T) {
	url := "https://www.mediafire.com/folder/a5st2zmorbatm/test"
	expectedKey := "a5st2zmorbatm"

	fkey := getFolderKey(url)
	if fkey != expectedKey {
		t.Errorf("1. getFolderKey error for url \"%s\": expected: %s, got:%s", url, expectedKey, fkey)
	}

	url = "https://www.mediafire.com/folder/a5st2zmorbatm/test/"
	expectedKey = "a5st2zmorbatm"

	fkey = getFolderKey(url)
	if fkey != expectedKey {
		t.Errorf("2. getFolderKey error for url \"%s\": expected: %s, got:%s", url, expectedKey, fkey)
	}
}

func TestMediafireFetchFolderContent(t *testing.T) {
	url := "https://www.mediafire.com/folder/a5st2zmorbatm/test"
	rootDirName := "test"

	expected := []*fileData{
		{
			filepath.Join(rootDirName),
			"1108148062423097515",
			"https://www.mediafire.com/file/j052b61nxp1hfop/1108148062423097515.mp3/file",
		},
		{
			filepath.Join(rootDirName),
			"NeapolitanFreshChannel - Cacio e maccarun",
			"https://www.mediafire.com/file/fpmgvsvpqsrsao4/NeapolitanFreshChannel_-_Cacio_e_maccarun.mp3/file",
		},
		{
			filepath.Join(rootDirName, "1"),
			"The Living Tombstone - My Ordinary Life Instrumental",
			"https://www.mediafire.com/file/24oavcjv6vxngj0/The_Living_Tombstone_-_My_Ordinary_Life_Instrumental.mp3/file",
		},
		{
			filepath.Join(rootDirName, "3"),
			"23343",
			"https://www.mediafire.com/file/cz7weo0q1fvbs95/23343.png/file",
		},
		{
			filepath.Join(rootDirName, "3"),
			"lady-gaga-bloody-mary-tiktok-remix-speed-up",
			"https://www.mediafire.com/file/56656tr32zman0s/lady-gaga-bloody-mary-tiktok-remix-speed-up.mp3/file",
		},
		{
			filepath.Join(rootDirName, "3", "3a"),
			"Animadrop - When a Champion Falls",
			"https://www.mediafire.com/file/u0jsjlf57455zmt/Animadrop_-_When_a_Champion_Falls.mp3/file",
		},
	}

	fk := getFolderKey(url)

	fd, err := fetchFolderContent(fk, rootDirName)
	if err != nil {
		t.Errorf("fetchFolderContent error: %s", err)
	}

	// length equality check
	if len(fd) != len(expected) {
		t.Errorf("fetchFolderContent error (lenght equality): expected: %v, got: %v", len(expected), len(fd))
	}

	// content equality check
	for i := 0; i < len(fd); i++ {
		if fd[i].Directory != expected[i].Directory {
			t.Errorf("fetchFolderContent error (directory equality): expected: %v, got: %v", expected[i].Directory, fd[i].Directory)
		}

		if fd[i].Filename != expected[i].Filename {
			t.Errorf("fetchFolderContent error (filename equality): expected: %v, got: %v", expected[i].Filename, fd[i].Filename)
		}

		if fd[i].Url != expected[i].Url {
			t.Errorf("fetchFolderContent error (url equality): expected: \"%v\", got: \"%v\"", expected[i].Url, fd[i].Url)
		}
	}
}

func TestMediafireFolderDownload(t *testing.T) {
	url := "https://www.mediafire.com/folder/a5st2zmorbatm/test"

	pwc, err := playwrightWrapper.UsePlaywright("chromium", false, playwrightWrapper.WithTimeout(0.0))
	if err != nil {
		t.Fatal(err)
	}
	defer pwc.Close()

	p, err := pwc.BrowserContext.NewPage()
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	_, err = p.Goto(url)
	if err != nil {
		t.Fatal(err)
	}

	albumName := "test"
	progress := int8(0)

	err = Mediafire(albumName, p, &progress)
	if err != nil {
		t.Fatal(err)
	}
}
