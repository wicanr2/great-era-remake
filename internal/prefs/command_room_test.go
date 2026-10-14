package prefs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndependentInterfaceAndArtMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs.json")
	p, err := Load(path)
	if err != nil || p.Layout != "command" || p.Theme != "modern" {
		t.Fatalf("fresh default %+v %v", p, err)
	}
	if err := os.WriteFile(path, []byte(`{"theme":"retro","message_time":3}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err = Load(path)
	if err != nil || p.Theme != "retro" || p.Layout != "command" || p.MessageTime != 3 {
		t.Fatalf("migration %+v %v", p, err)
	}
	for _, ui := range []string{"command", "original"} {
		for _, art := range []string{"modern", "retro"} {
			p.Layout = ui
			p.Theme = art
			if err := Save(path, p); err != nil {
				t.Fatal(err)
			}
			got, err := Load(path)
			if err != nil || got.Layout != ui || got.Theme != art {
				t.Fatal("coupled preferences")
			}
		}
	}
	p.Layout = "bogus"
	if err := Save(path, p); err == nil {
		t.Fatal("invalid layout accepted")
	}
}
