package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZeroArgumentResourcesAreBesideExecutableNotWorkingDirectory(t *testing.T) {
	for _, platform := range []string{"windows", "darwin", "linux"} {
		t.Run(platform, func(t *testing.T) {
			root := t.TempDir()
			exe := filepath.Join(root, "dsds.exe")
			resources := root
			if platform == "darwin" {
				exe = filepath.Join(root, "Contents/MacOS/dsds")
				resources = filepath.Join(root, "Contents/Resources")
			}
			if platform == "linux" {
				exe = filepath.Join(root, "usr/bin/dsds")
				resources = filepath.Join(root, "usr/share/great-era")
			}
			for _, marker := range []string{"data/game/WARPOS.DAT", "data/eten/STDFONT.15", "translations/zh-Hant/wording.json", "assets/music/modern/manifest.json"} {
				p := filepath.Join(resources, marker)
				if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte("fixture"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			p := startupPathsFor(exe, t.TempDir(), platform, func(k string) string { return "" }, func() (string, error) { return root, nil })
			if p.game != filepath.Join(resources, "data/game") || p.eten != filepath.Join(resources, "data/eten") || p.locale != filepath.Join(resources, "translations/zh-Hant") || p.music != filepath.Join(resources, "assets/music/modern") {
				t.Fatalf("paths=%+v", p)
			}
			if filepath.Dir(p.save) == resources {
				t.Fatal("save inside resources")
			}
		})
	}
}
