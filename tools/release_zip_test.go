package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReleaseZipIsOrderedReproducibleAndPreservesExecutableMode(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "GreatEraRemake")
	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("release note\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "dsds"), []byte("binary placeholder\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "RUNNING.md"), []byte("run\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	first := filepath.Join(root, "first.zip")
	second := filepath.Join(root, "second.zip")
	if err := run(source, first); err != nil {
		t.Fatal(err)
	}
	if err := run(source, second); err != nil {
		t.Fatal(err)
	}
	if err := verify(first, "GreatEraRemake/dsds"); err != nil {
		t.Fatalf("驗證剛建立的 ZIP：%v", err)
	}
	if err := verify(first, "GreatEraRemake/missing"); err == nil {
		t.Fatal("缺少必要發行入口的 ZIP 應拒絕")
	}
	firstBytes, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := os.ReadFile(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("相同輸入的 ZIP 不可重現")
	}

	reader, err := zip.OpenReader(first)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	want := []string{
		"GreatEraRemake/README.md",
		"GreatEraRemake/docs/",
		"GreatEraRemake/docs/RUNNING.md",
		"GreatEraRemake/dsds",
	}
	if len(reader.File) != len(want) {
		t.Fatalf("ZIP 項目數=%d，預期 %d", len(reader.File), len(want))
	}
	for i, file := range reader.File {
		if file.Name != want[i] {
			t.Fatalf("ZIP 順序 %d=%q，預期 %q", i, file.Name, want[i])
		}
		if !file.Modified.Equal(zipEpoch) {
			t.Fatalf("%s 的 ZIP 時間=%s，預期固定 epoch=%s", file.Name, file.Modified, zipEpoch)
		}
		if file.Name == "GreatEraRemake/dsds" && file.Mode().Perm()&0o111 == 0 {
			t.Fatalf("執行檔 mode 未保留：%s", file.Mode())
		}
	}
}
