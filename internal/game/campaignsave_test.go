package game

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"testing"
)

func TestCampaignContainerRejectsMalformedArchivesAndJSONShape(t *testing.T) {
	type state struct {
		Numbers [3]int
		Seed    uint32
	}
	good := state{Numbers: [3]int{1, 2, 3}, Seed: 0}
	encoded, err := EncodeCampaignSave(good, []byte{0, 255, 42})
	if err != nil {
		t.Fatal(err)
	}
	var got state
	legacy, err := DecodeCampaignSave(encoded, &got)
	if err != nil || got != good || !bytes.Equal(legacy, []byte{0, 255, 42}) {
		t.Fatalf("完整快照存讀失敗：%v", err)
	}
	for name, raw := range map[string]string{"short-array": `{"Numbers":[1,2],"Seed":0}`, "missing-field": `{"Numbers":[1,2,3]}`, "unknown-field": `{"Numbers":[1,2,3],"Seed":0,"Extra":1}`, "duplicate-field": `{"Numbers":[1,2,3],"Seed":0,"Seed":1}`, "null-seed": `{"Numbers":[1,2,3],"Seed":null}`} {
		t.Run(name, func(t *testing.T) {
			b, e := EncodeCampaignSave(json.RawMessage(raw), nil)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = DecodeCampaignSave(b, &got); e == nil {
				t.Fatal("破損 JSON 被接受")
			}
		})
	}
	for name, entries := range map[string][]string{"duplicate": {"manifest.json", "manifest.json", "state.json"}, "path": {"../state.json", "manifest.json"}, "unknown": {"manifest.json", "state.json", "future.json"}, "missing": {"state.json"}} {
		t.Run(name, func(t *testing.T) {
			var b bytes.Buffer
			z := zip.NewWriter(&b)
			for _, n := range entries {
				w, e := z.Create(n)
				if e != nil {
					t.Fatal(e)
				}
				w.Write([]byte(`{}`))
			}
			z.Close()
			if _, e := DecodeCampaignSave(b.Bytes(), &got); e == nil {
				t.Fatal("破損 archive 被接受")
			}
		})
	}
	if _, err := DecodeCampaignSave(make([]byte, CampaignSaveLimit+1), &got); err == nil {
		t.Fatal("超量容器被接受")
	}
}

func TestCampaignContainerVersionsRetainLegacyAndRejectUnknown(t *testing.T) {
	type snapshot struct{ Seed uint32 }
	for version := 1; version <= CampaignSaveCurrentVersion; version++ {
		raw, err := EncodeCampaignSaveVersion(snapshot{Seed: 42}, []byte{0, 255, 0, 17}, version)
		if err != nil {
			t.Fatal(err)
		}
		gotVersion, err := CampaignSaveVersion(raw)
		if err != nil || gotVersion != version {
			t.Fatal(gotVersion, err)
		}
		var state snapshot
		legacy, err := DecodeCampaignSaveVersion(raw, &state, version)
		if err != nil || state.Seed != 42 || !bytes.Equal(legacy, []byte{0, 255, 0, 17}) {
			t.Fatal("version or legacy source lost", version, err)
		}
		if _, err := DecodeCampaignSaveVersion(raw, &state, version%CampaignSaveCurrentVersion+1); err == nil {
			t.Fatal("mismatched DTO version accepted")
		}
	}
	for _, version := range []int{0, CampaignSaveCurrentVersion + 1} {
		if _, err := EncodeCampaignSaveVersion(snapshot{}, nil, version); err == nil {
			t.Fatal("unknown version accepted", version)
		}
	}
}
