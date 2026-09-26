package audio

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

// ModernTrackSource 是一首新編曲 Ogg/Vorbis 曲目及其循環點。
// LoopStart／LoopLength 使用解碼後「每聲道 sample」計數；零長度代表整首循環。
// 這裡只接受新增且有 provenance 的音檔，不讀取原版 MUS/TIM。
type ModernTrackSource struct {
	OGG        []byte
	LoopStart  int64
	LoopLength int64
}

// ModernManifest 是 assets/music/modern/manifest.json 的最小、可追溯格式。
type ModernManifest struct {
	Schema  int                            `json:"schema"`
	Version string                         `json:"version"`
	Tracks  map[string]ModernManifestTrack `json:"tracks"`
}

// ModernManifestTrack 描述檔名、循環 metadata 與音檔來源。作者／授權／雜湊
// 是執行期允許載入的必要欄位，避免把暫存或未授權音檔誤放進發行包。
type ModernManifestTrack struct {
	File       string `json:"file"`
	LoopStart  int64  `json:"loop_start"`
	LoopLength int64  `json:"loop_length"`
	Author     string `json:"author"`
	License    string `json:"license"`
	SHA256     string `json:"sha256"`
}

// ModernTrackSet 是 manifest 載入結果。缺少非 scene 曲目只產生警告，讓
// manager 沿用目前 cue； scene 缺失或任何 provenance／雜湊錯誤則整組停用。
type ModernTrackSet struct {
	Sources  map[Track]ModernTrackSource
	Warnings []string
}

var modernManifestTracks = map[string]Track{
	"modern_scene":    TrackScene,
	"modern_strategy": TrackStrategy,
	"modern_battle_a": TrackBattle1,
	"modern_battle_b": TrackBattle2,
	"modern_story":    TrackWall,
	"modern_final":    TrackFinal,
}

// LoadModernTracks 讀取並驗證一個 modern 音樂目錄。它不會掃描目錄或猜測
// 檔名，只接受 manifest 明列的六個 cue；讀不到 optional cue 時保留 warning。
func LoadModernTracks(dir string) (ModernTrackSet, error) {
	var result ModernTrackSet
	manifestData, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return result, fmt.Errorf("讀不到 modern manifest：%w", err)
	}
	var manifest ModernManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return result, fmt.Errorf("解析 modern manifest：%w", err)
	}
	if manifest.Schema != 1 {
		return result, fmt.Errorf("modern manifest schema=%d；只接受 schema=1", manifest.Schema)
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return result, errors.New("modern manifest 缺少 version")
	}
	if len(manifest.Tracks) == 0 {
		return result, errors.New("modern manifest 沒有 tracks")
	}

	result.Sources = make(map[Track]ModernTrackSource, len(manifest.Tracks))
	ids := make([]string, 0, len(manifest.Tracks))
	for id := range manifest.Tracks {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		meta := manifest.Tracks[id]
		track, ok := modernManifestTracks[id]
		if !ok {
			return ModernTrackSet{}, fmt.Errorf("modern manifest 含未知 cue %q", id)
		}
		if err := validateModernMeta(id, meta); err != nil {
			return ModernTrackSet{}, err
		}
		path := filepath.Join(dir, meta.File)
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			if errors.Is(readErr, os.ErrNotExist) && track != TrackScene {
				result.Warnings = append(result.Warnings, fmt.Sprintf("cue %s 缺少 %s，略過", id, meta.File))
				continue
			}
			return ModernTrackSet{}, fmt.Errorf("讀取 modern cue %s：%w", id, readErr)
		}
		digest := sha256.Sum256(data)
		gotHash := hex.EncodeToString(digest[:])
		if !strings.EqualFold(gotHash, meta.SHA256) {
			return ModernTrackSet{}, fmt.Errorf("modern cue %s SHA-256 不符：manifest=%s，實際=%s", id, meta.SHA256, gotHash)
		}
		result.Sources[track] = ModernTrackSource{
			OGG:        append([]byte(nil), data...),
			LoopStart:  meta.LoopStart,
			LoopLength: meta.LoopLength,
		}
	}
	if _, ok := result.Sources[TrackScene]; !ok {
		return ModernTrackSet{}, errors.New("modern manifest 缺少可載入的 modern_scene")
	}
	return result, nil
}

func validateModernMeta(id string, meta ModernManifestTrack) error {
	if meta.File == "" || filepath.Base(meta.File) != meta.File || strings.ContainsAny(meta.File, `/\\`) || meta.File == "." || meta.File == ".." {
		return fmt.Errorf("modern cue %s 的 file 必須是同目錄 .ogg 檔名", id)
	}
	if strings.ToLower(filepath.Ext(meta.File)) != ".ogg" {
		return fmt.Errorf("modern cue %s 不是 .ogg：%q", id, meta.File)
	}
	if meta.LoopStart < 0 || meta.LoopLength < 0 {
		return fmt.Errorf("modern cue %s 的 loop metadata 不可為負數", id)
	}
	if strings.TrimSpace(meta.Author) == "" || strings.TrimSpace(meta.License) == "" {
		return fmt.Errorf("modern cue %s 必須記錄 author 與 license", id)
	}
	decoded, err := hex.DecodeString(meta.SHA256)
	if err != nil || len(decoded) != sha256.Size {
		return fmt.Errorf("modern cue %s 的 sha256 必須是 64 位十六進位字串", id)
	}
	return nil
}

// NewModernTracks 驗證初始 Ogg，依其取樣率建立 Ebiten audio.Context；其餘
// 曲目延遲到切換時解碼。解碼器是 Ebiten 內建的純 Go Vorbis 路徑，不使用 cgo。
func NewModernTracks(sources map[Track]ModernTrackSource, initial Track) (*Manager, error) {
	if len(sources) == 0 {
		return nil, errors.New("audio: 至少需要一首 modern 曲目")
	}
	if initial == "" {
		return nil, errors.New("audio: 初始曲目不可為空")
	}
	source, ok := sources[initial]
	if !ok {
		return nil, fmt.Errorf("audio: 缺少初始 modern 曲目 %q", initial)
	}
	stream, err := decodeModernStream(source.OGG, 0)
	if err != nil {
		return nil, fmt.Errorf("audio: 驗證 modern 曲目 %q：%w", initial, err)
	}
	if stream.SampleRate() <= 0 {
		return nil, fmt.Errorf("audio: modern 曲目 %q 沒有有效取樣率", initial)
	}
	ctx := ebitenaudio.NewContext(stream.SampleRate())
	copySources := make(map[Track]ModernTrackSource, len(sources))
	for track, value := range sources {
		copySources[track] = ModernTrackSource{
			OGG:        append([]byte(nil), value.OGG...),
			LoopStart:  value.LoopStart,
			LoopLength: value.LoopLength,
		}
	}
	return &Manager{
		mode:               ModeModern,
		ctx:                ebitenContext{ctx: ctx},
		modernSources:      copySources,
		proceduralFallback: true,
		current:            initial,
		sampleRateValue:    stream.SampleRate(),
		volume:             0.7,
	}, nil
}

func decodeModernStream(data []byte, sampleRate int) (*vorbis.Stream, error) {
	if len(data) == 0 {
		return nil, errors.New("Ogg bytes 為空")
	}
	if sampleRate > 0 {
		return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	}
	return vorbis.DecodeWithoutResampling(bytes.NewReader(data))
}

// loopReader 把已解碼的 Ogg 串流限制在宣告的 loop 範圍，讀到尾端後 seek
// 回 loop 起點。Ebiten player 會在背景讀取它；Stop／Close 時會中止 player。
type loopReader struct {
	src   io.ReadSeeker
	start int64
	end   int64
	pos   int64
}

func newLoopReader(stream *vorbis.Stream, source ModernTrackSource) (io.Reader, error) {
	if stream == nil {
		return nil, errors.New("Ogg stream 為空")
	}
	length := stream.Length()
	if length <= 0 {
		if source.LoopStart != 0 || source.LoopLength != 0 {
			return nil, errors.New("Ogg 解碼長度未知，不能驗證 loop metadata")
		}
		return stream, nil
	}
	const stereoI16BytesPerSample = int64(4)
	if source.LoopLength == 0 {
		if source.LoopStart != 0 {
			return nil, errors.New("loop_length=0 時 loop_start 必須為 0")
		}
		return newFullLoopReader(stream, length)
	}
	if source.LoopStart > length/stereoI16BytesPerSample ||
		source.LoopLength > length/stereoI16BytesPerSample {
		return nil, fmt.Errorf("loop_start=%d loop_length=%d 超出解碼長度 %d bytes", source.LoopStart, source.LoopLength, length)
	}
	start := source.LoopStart * stereoI16BytesPerSample
	loopLength := source.LoopLength * stereoI16BytesPerSample
	if start < 0 || loopLength <= 0 || start >= length || loopLength > length-start {
		return nil, fmt.Errorf("loop_start=%d loop_length=%d 超出解碼長度 %d bytes", source.LoopStart, source.LoopLength, length)
	}
	if _, err := stream.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}
	return &loopReader{src: stream, start: start, end: start + loopLength, pos: start}, nil
}

func newFullLoopReader(stream *vorbis.Stream, length int64) (io.Reader, error) {
	if _, err := stream.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return &loopReader{src: stream, start: 0, end: length, pos: 0}, nil
}

func (r *loopReader) Read(p []byte) (int, error) {
	if r == nil || r.src == nil {
		return 0, errors.New("audio: loop reader 為空")
	}
	if len(p) == 0 {
		return 0, nil
	}
	n := 0
	for n < len(p) {
		if r.pos >= r.end {
			if _, err := r.src.Seek(r.start, io.SeekStart); err != nil {
				return n, err
			}
			r.pos = r.start
		}
		want := int64(len(p) - n)
		if remain := r.end - r.pos; want > remain {
			want = remain
		}
		if want <= 0 {
			continue
		}
		readN, err := r.src.Read(p[n : n+int(want)])
		if readN > 0 {
			n += readN
			r.pos += int64(readN)
		}
		if err != nil && err != io.EOF {
			return n, err
		}
		if readN == 0 && err == io.EOF {
			r.pos = r.end
		}
		if readN == 0 && err == nil {
			return n, io.ErrNoProgress
		}
	}
	return n, nil
}
