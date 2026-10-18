package game

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

const CampaignSaveLimit = 8 << 20

const CampaignSaveCurrentVersion = 4

type campaignManifest struct {
	Format       string
	Version      int
	StateSHA256  string
	LegacySHA256 string
}

func campaignDigest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

// EncodeCampaignSave 只編碼重製版容器，原始 DT1 保留為不變的匯入來源。
func EncodeCampaignSave(state any, legacy []byte) ([]byte, error) {
	return EncodeCampaignSaveVersion(state, legacy, 1)
}

func EncodeCampaignSaveVersion(state any, legacy []byte, version int) ([]byte, error) {
	if version < 1 || version > CampaignSaveCurrentVersion {
		return nil, fmt.Errorf("game: 不支援的保存版本")
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	if len(payload)+len(legacy) > CampaignSaveLimit {
		return nil, fmt.Errorf("game: 保存內容超過容量上限")
	}
	m := campaignManifest{Format: "great-era-save", Version: version, StateSHA256: campaignDigest(payload)}
	if len(legacy) > 0 {
		m.LegacySHA256 = campaignDigest(legacy)
	}
	manifest, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	z := zip.NewWriter(&out)
	for _, item := range []struct {
		name string
		data []byte
	}{{"manifest.json", manifest}, {"state.json", payload}, {"legacy.dt1", legacy}} {
		if item.name == "legacy.dt1" && len(item.data) == 0 {
			continue
		}
		w, e := z.Create(item.name)
		if e != nil {
			return nil, e
		}
		if _, e = w.Write(item.data); e != nil {
			return nil, e
		}
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	if out.Len() > CampaignSaveLimit {
		return nil, fmt.Errorf("game: 保存封包超過容量上限")
	}
	return out.Bytes(), nil
}

// DecodeCampaignSave 不解壓到磁碟。格式、完整性與完整 JSON 形狀全數驗證後才解碼。
// 呼叫端仍須以獨立 DTO 驗證語意，全部成功後才能替換現行遊戲。
func DecodeCampaignSave(data []byte, state any) ([]byte, error) {
	return DecodeCampaignSaveVersion(data, state, 1)
}

// CampaignSaveVersion selects a DTO only; full archive and payload validation
// still happens in DecodeCampaignSaveVersion before a caller applies state.
func CampaignSaveVersion(data []byte) (int, error) {
	if len(data) > CampaignSaveLimit {
		return 0, fmt.Errorf("game: 保存封包超過容量上限")
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return 0, err
	}
	var manifest []byte
	for _, f := range z.File {
		if f.Name != "manifest.json" {
			continue
		}
		if manifest != nil || f.UncompressedSize64 > 4096 {
			return 0, fmt.Errorf("game: 保存 manifest 不合法")
		}
		r, e := f.Open()
		if e != nil {
			return 0, e
		}
		manifest, e = io.ReadAll(io.LimitReader(r, 4097))
		closed := r.Close()
		if e != nil {
			return 0, e
		}
		if closed != nil {
			return 0, closed
		}
	}
	if len(manifest) == 0 || len(manifest) > 4096 {
		return 0, fmt.Errorf("game: 保存缺少有效 manifest")
	}
	var m campaignManifest
	if err := decodeCampaignJSON(manifest, &m); err != nil {
		return 0, err
	}
	if m.Format != "great-era-save" || m.Version < 1 || m.Version > CampaignSaveCurrentVersion {
		return 0, fmt.Errorf("game: 不支援的保存格式或版本")
	}
	return m.Version, nil
}

func DecodeCampaignSaveVersion(data []byte, state any, version int) ([]byte, error) {
	if version < 1 || version > CampaignSaveCurrentVersion {
		return nil, fmt.Errorf("game: 不支援的保存版本")
	}
	if len(data) > CampaignSaveLimit {
		return nil, fmt.Errorf("game: 保存封包超過容量上限")
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("game: 無法讀取保存容器：%w", err)
	}
	items := make(map[string][]byte)
	total := uint64(0)
	for _, f := range z.File {
		if f.Name != "manifest.json" && f.Name != "state.json" && f.Name != "legacy.dt1" {
			return nil, fmt.Errorf("game: 未知保存項目 %q", f.Name)
		}
		if _, found := items[f.Name]; found {
			return nil, fmt.Errorf("game: 重複保存項目 %q", f.Name)
		}
		if f.UncompressedSize64 > CampaignSaveLimit || total+f.UncompressedSize64 > CampaignSaveLimit {
			return nil, fmt.Errorf("game: 解壓內容超過容量上限")
		}
		total += f.UncompressedSize64
		r, e := f.Open()
		if e != nil {
			return nil, e
		}
		b, e := io.ReadAll(io.LimitReader(r, CampaignSaveLimit+1))
		closeErr := r.Close()
		if e != nil {
			return nil, e
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if uint64(len(b)) != f.UncompressedSize64 {
			return nil, fmt.Errorf("game: 保存項目大小不一致")
		}
		items[f.Name] = b
	}
	if len(items["manifest.json"]) == 0 || len(items["state.json"]) == 0 {
		return nil, fmt.Errorf("game: 保存容器缺少必要項目")
	}
	var m campaignManifest
	if err := decodeCampaignJSON(items["manifest.json"], &m); err != nil {
		return nil, err
	}
	if m.Format != "great-era-save" || m.Version != version {
		return nil, fmt.Errorf("game: 不支援的保存格式或版本")
	}
	legacy, hasLegacy := items["legacy.dt1"]
	if campaignDigest(items["state.json"]) != m.StateSHA256 || hasLegacy != (m.LegacySHA256 != "") || (hasLegacy && campaignDigest(legacy) != m.LegacySHA256) {
		return nil, fmt.Errorf("game: 保存容器雜湊不一致")
	}
	if err := decodeCampaignJSON(items["state.json"], state); err != nil {
		return nil, err
	}
	return append([]byte(nil), legacy...), nil
}

func decodeCampaignJSON(data []byte, dst any) error {
	t := reflect.TypeOf(dst)
	if t == nil || t.Kind() != reflect.Pointer || reflect.ValueOf(dst).IsNil() {
		return fmt.Errorf("game: 保存解碼需要有效指標")
	}
	if err := campaignUniqueJSON(json.NewDecoder(bytes.NewReader(data)), 0); err != nil {
		return err
	}
	// json.Unmarshal 對短陣列補零，也接受缺欄。保存格式須明確拒絕這些情況。
	if err := campaignJSONShape(data, t.Elem()); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return fmt.Errorf("game: 保存 JSON：%w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("game: 保存 JSON 含多餘內容")
	}
	return nil
}

func campaignUniqueJSON(d *json.Decoder, depth int) error {
	if depth > 64 {
		return fmt.Errorf("game: 保存 JSON 巢狀過深")
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delim != '{' && delim != '[' {
		return fmt.Errorf("game: 保存 JSON 結構不合法")
	}
	seen := map[string]bool{}
	for d.More() {
		if delim == '{' {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return fmt.Errorf("game: 保存 JSON 重複或不合法欄位")
			}
			seen[name] = true
		}
		if err := campaignUniqueJSON(d, depth+1); err != nil {
			return err
		}
	}
	end, err := d.Token()
	if err != nil {
		return err
	}
	if (delim == '{' && end != json.Delim('}')) || (delim == '[' && end != json.Delim(']')) {
		return fmt.Errorf("game: 保存 JSON 結尾不合法")
	}
	return nil
}

func campaignJSONShape(data json.RawMessage, t reflect.Type) error {
	if t.Kind() == reflect.Pointer {
		if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
			return nil
		}
		return campaignJSONShape(data, t.Elem())
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) && t.Kind() != reflect.Map && t.Kind() != reflect.Slice {
		return fmt.Errorf("game: 必填保存值不能為 null")
	}
	switch t.Kind() {
	case reflect.Struct:
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(data, &obj); err != nil || obj == nil {
			return fmt.Errorf("game: 保存物件不合法")
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				if err := campaignJSONShape(data, f.Type); err != nil {
					return err
				}
				continue
			}
			b, ok := obj[f.Name]
			if !ok {
				if strings.Contains(","+f.Tag.Get("json")+",", ",omitempty,") {
					continue
				}
				return fmt.Errorf("game: 保存缺少欄位 %s", f.Name)
			}
			if err := campaignJSONShape(b, f.Type); err != nil {
				return fmt.Errorf("%s: %w", f.Name, err)
			}
		}
	case reflect.Array, reflect.Slice:
		if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
			return nil
		}
		var list []json.RawMessage
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		if t.Kind() == reflect.Array && len(list) != t.Len() {
			return fmt.Errorf("game: 保存陣列需要 %d 筆，實得 %d", t.Len(), len(list))
		}
		for _, b := range list {
			if err := campaignJSONShape(b, t.Elem()); err != nil {
				return err
			}
		}
	case reflect.Map:
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		for _, b := range obj {
			if err := campaignJSONShape(b, t.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}
