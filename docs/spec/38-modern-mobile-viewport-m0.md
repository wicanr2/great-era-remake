# SPEC-38：Modern 高解析行動裝置 viewport（M0）

狀態：**READY／純 Go adapter 已實作，Android binding 未宣稱完成**　日期：2026-08-11

## 目的

讓 Android／平板日後只需接平台生命週期與 `ebitenmobile`，就能把 Modern 的
1280×720 設計畫布安全地放入實際 Surface。安全區、letterbox、density 與 48 dp
觸控命中必須在不依賴 Android／cgo 的純 Go 層先固定，避免平台 adapter 重複實作
命中邏輯。

## 契約

- `mobile.NewViewport` 以實際 Surface 像素、density 與系統安全區建立等比 viewport；
  不拉伸、不把直向 Surface 偷換成橫向遊戲畫布。
- `SurfaceToDesign` 拒絕安全區／letterbox 外的點，成功時回傳 Modern 設計座標；
  後續一律交給既有 `actions.Hit` 與 `layout.Rect`。
- `DesignToSurface` 只供焦點框／debug overlay，不能成為第二份規則命中表。
- `TouchTarget` 只擴大不可見的觸控命中區，依 density／scale 保證至少 48 dp；視覺
  卡片與桌面 pointer 矩形不被改寫。
- 這層不匯入 Ebiten、Android、NDK 或 cgo；`ebitenmobile bind`、背景恢復、系統返回、
  真機音訊與 APK／AAB 仍屬 H3/H5 發行 gate。

## 完成證據

- `internal/ui/mobile/viewport.go`：viewport、座標反算、safe inset 與 48 dp 擴張。
- `internal/ui/mobile/viewport_test.go`：1920×1080 round-trip、letterbox 拒絕、瀏海
  安全區、density 3 的 48 dp 與非法輸入測試。
- `docs/design/33-high-resolution-modern-android.md`：H3 Android binding 與實機閘門。

## 邊界

M0 只證明平台無關的幾何／輸入契約；不證明 Android binding 可以在目前環境建出，也不
證明實機旋轉、背景切回、音訊裝置或商店包完成。這些證據不能由桌面 Xvfb 測試取代。
