package i18n

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPortraitCatalogVerifiesAndUsesCanonicalPerson(t *testing.T) {
	shared, hash := writePortraitFixture(t)
	record := validPortraitRecord(hash)
	writePortraitRegistry(t, shared, []portraitRecord{record})
	people := portraitTestPeople()
	catalog, err := LoadPortraitCatalog(shared, people)
	if err != nil {
		t.Fatal(err)
	}
	portrait, ok := catalog.PortraitFor(2) // #2 是 #1 的同人別名，應回到 canonical #1。
	if !ok || portrait.PersonID != 1 || portrait.Image.Bounds().Dx() != 8 || portrait.Image.Bounds().Dy() != 12 {
		t.Fatalf("canonical 查詢結果不符：ok=%v portrait=%+v", ok, portrait)
	}
	if _, ok := catalog.PortraitFor(99); ok {
		t.Fatal("沒有登錄的人物不得自動配圖")
	}
}

func TestLoadPortraitCatalogRejectsUnsafeOrUnverifiedRecords(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*portraitRecord)
	}{
		{"未知授權", func(r *portraitRecord) { r.Rights.Kind = "cc-by-nc-4.0" }},
		{"缺少署名", func(r *portraitRecord) { r.Rights.Attribution = "" }},
		{"雜湊不符", func(r *portraitRecord) {
			r.AssetSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}},
		{"跨目錄路徑", func(r *portraitRecord) { r.Asset = "../outside.jpg" }},
		{"未確認身分", func(r *portraitRecord) { r.Identity.Verification = "hypothesis" }},
		{"非 canonical 人物", func(r *portraitRecord) { r.PersonID, r.Identity.Name = 2, "別名" }},
		{"排除槽位", func(r *portraitRecord) { r.PersonID, r.Identity.Name = 274, "無省長" }},
		{"非法裁切", func(r *portraitRecord) { r.Crop.Width = 1.1 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			shared, hash := writePortraitFixture(t)
			record := validPortraitRecord(hash)
			tc.mutate(&record)
			writePortraitRegistry(t, shared, []portraitRecord{record})
			if _, err := LoadPortraitCatalog(shared, portraitTestPeople()); err == nil {
				t.Fatal("不安全或未確認肖像必須被拒絕")
			}
		})
	}
}

func portraitTestPeople() *PeopleDB {
	return &PeopleDB{
		people: map[int]*Person{
			1:   {ID: 1, NameInGame: "測試人物"},
			2:   {ID: 2, NameInGame: "別名"},
			274: {ID: 274, NameInGame: "無省長"},
		},
		canonical: map[int]int{1: 1, 2: 1, 274: 274},
	}
}

func writePortraitFixture(t *testing.T) (string, string) {
	t.Helper()
	shared := t.TempDir()
	dir := filepath.Join(shared, "portraits")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 8, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 20), G: uint8(y * 12), B: 90, A: 255})
		}
	}
	path := filepath.Join(dir, "0001-test.jpg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	return shared, hex.EncodeToString(sum[:])
}

func validPortraitRecord(hash string) portraitRecord {
	return portraitRecord{
		PersonID: 1, Status: "available", Asset: "portraits/0001-test.jpg", AssetSHA256: hash,
		Identity: portraitIdentity{Name: "測試人物", Verification: "confirmed", ReferenceURL: "https://example.test/person/1"},
		Rights: portraitRights{Kind: "cc-by-4.0", LicenseURL: "https://creativecommons.org/licenses/by/4.0/",
			Author: "Test Author", Attribution: "Test Author, CC BY 4.0", SourcePage: "https://example.test/file/1", RetrievedAt: "2026-08-12"},
		Review: portraitReview{Identity: "confirmed", Rights: "confirmed", ReviewedAt: "2026-08-12"},
		Crop:   PortraitCrop{X: 0, Y: 0, Width: 1, Height: 1},
	}
}

func writePortraitRegistry(t *testing.T, shared string, records []portraitRecord) {
	t.Helper()
	b, err := json.Marshal(portraitRegistry{SchemaVersion: "1", Portraits: records})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shared, portraitRegistryFile), b, 0o644); err != nil {
		t.Fatal(err)
	}
}
