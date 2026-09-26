package i18n

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const portraitRegistryFile = "portraits.json"

// Portrait 是已通過身分、權利與檔案完整性檢查的非原版肖像。它獨立於 Person，
// 因為圖像來源與授權不隨 UI 語系改變，且絕不能被誤當成原版將領資料。
type Portrait struct {
	PersonID    int
	Image       image.Image
	Asset       string
	Attribution string
	LicenseKind string
	LicenseURL  string
	SourcePage  string
	RetrievedAt string
	Crop        PortraitCrop
}

// PortraitCrop 使用來源影像寬高正規化的半開矩形；renderer 依它裁切而不修改來源檔。
type PortraitCrop struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// PortraitCatalog 把 canonical 人物 id 映到可畫的離線肖像。缺少條目是正常情況，
// 呼叫端必須顯示中性檔案卡，而不能找另一張圖代替。
type PortraitCatalog struct {
	portraits map[int]Portrait
	canonical map[int]int
}

type portraitRegistry struct {
	SchemaVersion string           `json:"schema_version"`
	Portraits     []portraitRecord `json:"portraits"`
}

type portraitRecord struct {
	PersonID    int              `json:"person_id"`
	Status      string           `json:"status"`
	Asset       string           `json:"asset"`
	AssetSHA256 string           `json:"asset_sha256"`
	Identity    portraitIdentity `json:"identity"`
	Rights      portraitRights   `json:"rights"`
	Review      portraitReview   `json:"review"`
	Crop        PortraitCrop     `json:"crop"`
}

type portraitIdentity struct {
	Name         string `json:"name"`
	Verification string `json:"verification"`
	ReferenceURL string `json:"reference_url"`
}

type portraitRights struct {
	Kind        string `json:"kind"`
	LicenseURL  string `json:"license_url"`
	Author      string `json:"author"`
	Attribution string `json:"attribution"`
	SourcePage  string `json:"source_page"`
	RetrievedAt string `json:"retrieved_at"`
}

type portraitReview struct {
	Identity   string `json:"identity"`
	Rights     string `json:"rights"`
	ReviewedAt string `json:"reviewed_at"`
}

// LoadPortraitCatalog 讀取語系無關的肖像登錄，並在載入時完成所有 fail-closed gate。
// registry 壞掉時回錯而非回傳部分成功資料；app 可退回中性檔案卡，但 release gate
// 不可接受不完整登錄。
func LoadPortraitCatalog(sharedDir string, people *PeopleDB) (*PortraitCatalog, error) {
	if people == nil {
		return nil, fmt.Errorf("i18n: 肖像登錄需要人物資料")
	}
	data, err := os.ReadFile(filepath.Join(sharedDir, portraitRegistryFile))
	if err != nil {
		return nil, fmt.Errorf("i18n: 讀取肖像登錄：%w", err)
	}
	var registry portraitRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("i18n: 解析肖像登錄：%w", err)
	}
	if registry.SchemaVersion != "1" {
		return nil, fmt.Errorf("i18n: 肖像登錄 schema_version=%q，預期 1", registry.SchemaVersion)
	}
	resolvedRoot, err := filepath.EvalSymlinks(sharedDir)
	if err != nil {
		return nil, fmt.Errorf("i18n: 解析肖像根目錄：%w", err)
	}
	catalog := &PortraitCatalog{
		portraits: make(map[int]Portrait, len(registry.Portraits)),
		canonical: make(map[int]int, len(people.canonical)),
	}
	for id, canonical := range people.canonical {
		catalog.canonical[id] = canonical
	}
	for _, record := range registry.Portraits {
		portrait, err := loadPortraitRecord(resolvedRoot, people, record)
		if err != nil {
			return nil, err
		}
		if _, exists := catalog.portraits[portrait.PersonID]; exists {
			return nil, fmt.Errorf("i18n: 肖像登錄人物 id=%d 重複", portrait.PersonID)
		}
		catalog.portraits[portrait.PersonID] = portrait
	}
	return catalog, nil
}

func loadPortraitRecord(sharedRoot string, people *PeopleDB, record portraitRecord) (Portrait, error) {
	if record.Status != "available" {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d status 必須為 available", record.PersonID)
	}
	person := people.people[record.PersonID]
	if person == nil || record.PersonID == 274 {
		return Portrait{}, fmt.Errorf("i18n: 肖像指向不存在或排除人物 id=%d", record.PersonID)
	}
	if canonical := people.canonical[record.PersonID]; canonical != record.PersonID {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d 不是 canonical id=%d", record.PersonID, canonical)
	}
	if record.Identity.Name != person.NameInGame || record.Identity.Verification != "confirmed" ||
		!validHTTPSURL(record.Identity.ReferenceURL) {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d 身分查證不完整", record.PersonID)
	}
	if !allowedPortraitLicense(record.Rights.Kind) || !validHTTPSURL(record.Rights.LicenseURL) ||
		!validHTTPSURL(record.Rights.SourcePage) || strings.TrimSpace(record.Rights.Author) == "" ||
		strings.TrimSpace(record.Rights.Attribution) == "" || !validDate(record.Rights.RetrievedAt) {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d 權利資料不完整或不允許", record.PersonID)
	}
	if record.Review.Identity != "confirmed" || record.Review.Rights != "confirmed" || !validDate(record.Review.ReviewedAt) {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d 審核狀態不完整", record.PersonID)
	}
	if !validPortraitCrop(record.Crop) {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d crop 無效", record.PersonID)
	}
	assetPath, assetName, err := portraitAssetPath(sharedRoot, record.Asset)
	if err != nil {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d：%w", record.PersonID, err)
	}
	if !validLowerSHA256(record.AssetSHA256) {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d SHA-256 格式無效", record.PersonID)
	}
	info, err := os.Lstat(assetPath)
	if err != nil {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d 讀不到圖檔：%w", record.PersonID, err)
	}
	if !info.Mode().IsRegular() {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d 圖檔不是一般檔案", record.PersonID)
	}
	data, err := os.ReadFile(assetPath)
	if err != nil {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d 讀取圖檔：%w", record.PersonID, err)
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != record.AssetSHA256 {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d 圖檔 SHA-256 不符", record.PersonID)
	}
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil || format != "jpeg" || img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		return Portrait{}, fmt.Errorf("i18n: 肖像人物 id=%d 不是有效 JPEG", record.PersonID)
	}
	return Portrait{
		PersonID: record.PersonID, Image: img, Asset: assetName,
		Attribution: record.Rights.Attribution, LicenseKind: record.Rights.Kind,
		LicenseURL: record.Rights.LicenseURL, SourcePage: record.Rights.SourcePage,
		RetrievedAt: record.Rights.RetrievedAt, Crop: record.Crop,
	}, nil
}

// PortraitFor 按 canonical 人物 id 查圖；不存在條目是正常資料狀態。
func (c *PortraitCatalog) PortraitFor(personID int) (Portrait, bool) {
	if c == nil {
		return Portrait{}, false
	}
	canonical, ok := c.canonical[personID]
	if !ok {
		return Portrait{}, false
	}
	p, ok := c.portraits[canonical]
	return p, ok
}

func allowedPortraitLicense(kind string) bool {
	switch kind {
	case "public-domain", "cc0-1.0", "cc-by-4.0", "cc-by-sa-4.0":
		return true
	default:
		return false
	}
}

func validPortraitCrop(c PortraitCrop) bool {
	return c.X >= 0 && c.Y >= 0 && c.Width > 0 && c.Height > 0 &&
		c.X+c.Width <= 1 && c.Y+c.Height <= 1
}

func validHTTPSURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Scheme == "https" && u.Host != ""
}

func validDate(raw string) bool {
	_, err := time.Parse("2006-01-02", raw)
	return err == nil
}

func validLowerSHA256(raw string) bool {
	if len(raw) != sha256.Size*2 || raw != strings.ToLower(raw) {
		return false
	}
	_, err := hex.DecodeString(raw)
	return err == nil
}

func portraitAssetPath(sharedRoot, asset string) (string, string, error) {
	if asset == "" {
		return "", "", fmt.Errorf("肖像 asset 不可空白")
	}
	clean := filepath.Clean(filepath.FromSlash(asset))
	rel := filepath.ToSlash(clean)
	if filepath.IsAbs(clean) || rel == "." || rel == ".." || strings.HasPrefix(rel, "../") ||
		!strings.HasPrefix(rel, "portraits/") {
		return "", "", fmt.Errorf("肖像 asset 必須是 portraits/ 下的相對路徑")
	}
	ext := strings.ToLower(filepath.Ext(clean))
	if ext != ".jpg" && ext != ".jpeg" {
		return "", "", fmt.Errorf("肖像 asset 必須是 JPEG")
	}
	path := filepath.Join(sharedRoot, clean)
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		if relToRoot, relErr := filepath.Rel(sharedRoot, resolved); relErr != nil || relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
			return "", "", fmt.Errorf("肖像 asset 解析後離開 shared 根目錄")
		}
	}
	return path, filepath.ToSlash(clean), nil
}
