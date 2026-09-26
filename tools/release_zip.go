// release_zip 以標準函式庫建立可重現的 ZIP 發行檔。
//
// 它只由 tools/package.sh 在 Docker 發行容器中呼叫，避免依賴主機 zip
// 工具或把未鎖定的封裝程式帶進專案。第一個引數是來源目錄，第二個引數是
// 輸出 ZIP；檔案時間固定為 ZIP 格式可表示的最早時間，保留可執行檔 mode。
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var zipEpoch = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

func main() {
	if len(os.Args) == 4 && os.Args[1] == "verify" {
		if err := verify(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, "release_zip verify:", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "用法：release_zip <來源目錄> <輸出.zip>；或 release_zip verify <輸出.zip> <必要項目>")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "release_zip:", err)
		os.Exit(1)
	}
}

// verify 讀取 ZIP 的每個檔案，以標準函式庫校驗 CRC 與路徑安全性，並確認
// 發行入口存在。它也是 package.sh 的封裝後獨立檢查，不只相信壓縮指令成功。
func verify(archive, required string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return fmt.Errorf("開啟 ZIP：%w", err)
	}
	defer reader.Close()
	if required == "" {
		return fmt.Errorf("必要項目不可為空")
	}
	found := false
	for _, file := range reader.File {
		name := file.Name
		clean := pathpkg.Clean(name)
		if name == "" || strings.HasPrefix(name, "/") || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
			return fmt.Errorf("ZIP 含不安全路徑：%q", name)
		}
		if !file.FileInfo().IsDir() {
			ext := strings.ToLower(pathpkg.Ext(name))
			switch ext {
			case ".tpc", ".15", ".rgb", ".mus", ".tim", ".dat", ".glb", ".gtb", ".dt1", ".dt2", ".sav", ".cps", ".bgi", ".ogg", ".wav", ".mid", ".midi":
				return fmt.Errorf("ZIP 含禁止的原版／衍生素材：%s", name)
			}
			in, openErr := file.Open()
			if openErr != nil {
				return fmt.Errorf("開啟 %s：%w", name, openErr)
			}
			_, copyErr := io.Copy(io.Discard, in)
			closeErr := in.Close()
			if copyErr != nil {
				return fmt.Errorf("讀取 %s：%w", name, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("關閉 %s：%w", name, closeErr)
			}
		}
		if name == required {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("ZIP 缺少必要項目：%s", required)
	}
	return nil
}

func run(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("讀取來源目錄：%w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("來源不是目錄：%s", source)
	}
	entries := make([]string, 0, 64)
	if err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return nil
		}
		entries = append(entries, path)
		return nil
	}); err != nil {
		return fmt.Errorf("列舉來源：%w", err)
	}
	sort.Strings(entries)

	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("建立輸出：%w", err)
	}
	defer out.Close()
	writer := zip.NewWriter(out)
	for _, path := range entries {
		if err := add(writer, source, path); err != nil {
			writer.Close()
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("完成 ZIP：%w", err)
	}
	return nil
}

func add(writer *zip.Writer, root, path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("讀取 %s：%w", path, err)
	}
	rel, err := filepath.Rel(filepath.Dir(root), path)
	if err != nil {
		return fmt.Errorf("取得相對路徑：%w", err)
	}
	name := filepath.ToSlash(rel)
	if info.IsDir() {
		name += "/"
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("建立 %s header：%w", path, err)
	}
	header.Name = name
	header.Method = zip.Deflate
	header.Modified = zipEpoch
	header.SetMode(info.Mode())
	if info.Mode()&os.ModeSymlink != 0 {
		target, targetErr := os.Readlink(path)
		if targetErr != nil {
			return fmt.Errorf("讀取符號連結 %s：%w", path, targetErr)
		}
		w, createErr := writer.CreateHeader(header)
		if createErr != nil {
			return fmt.Errorf("加入 %s：%w", path, createErr)
		}
		if _, writeErr := io.WriteString(w, target); writeErr != nil {
			return fmt.Errorf("寫入符號連結 %s：%w", path, writeErr)
		}
		return nil
	}
	w, err := writer.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("加入 %s：%w", path, err)
	}
	if info.IsDir() {
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("拒絕非一般檔案：%s (%s)", path, strings.TrimSpace(info.Mode().String()))
	}
	in, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("開啟 %s：%w", path, err)
	}
	defer in.Close()
	if _, err := io.Copy(w, in); err != nil {
		return fmt.Errorf("寫入 %s：%w", path, err)
	}
	return nil
}
