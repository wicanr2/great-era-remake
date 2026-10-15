package main

import (
	"os"
	"path/filepath"
	"runtime"
)

type startupPaths struct{ game, eten, locale, music, save string }

func resourcePath(roots []string, relative, marker string) string {
	for _, root := range roots {
		p := filepath.Join(root, relative)
		if st, err := os.Stat(filepath.Join(p, marker)); err == nil && !st.IsDir() {
			return p
		}
	}
	return relative
}

func defaultStartupPaths() startupPaths {
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	return startupPathsFor(exe, cwd, runtime.GOOS, os.Getenv, os.UserHomeDir)
}

func startupPathsFor(exe, cwd, platform string, env func(string) string, home func() (string, error)) startupPaths {
	dir := filepath.Dir(exe)
	roots := []string{dir, filepath.Join(dir, "../Resources"), filepath.Join(dir, "../share/great-era"), cwd}
	p := startupPaths{
		game:   resourcePath(roots, "data/game", "WARPOS.DAT"),
		eten:   resourcePath(roots, "data/eten", "STDFONT.15"),
		locale: resourcePath(roots, "translations/zh-Hant", "wording.json"),
		music:  resourcePath(roots, "assets/music/modern", "manifest.json"),
	}
	if p.game == "data/game" {
		p.game = filepath.Join(cwd, "workplace/orig/game")
	}
	if p.eten == "data/eten" {
		p.eten = filepath.Join(cwd, "workplace/eten")
	}
	h, _ := home()
	base := env("XDG_DATA_HOME")
	if base == "" {
		base = filepath.Join(h, ".local/share")
	}
	if platform == "darwin" {
		base = filepath.Join(h, "Library/Application Support")
	}
	if platform == "windows" {
		base = env("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(h, "AppData/Local")
		}
	}
	p.save = filepath.Join(base, "GreatEraRemake/Saves/SAVE(1).DT1")
	return p
}
