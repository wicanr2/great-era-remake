# RE-34 指令 15 「訊息時間」

狀態：**CONFIRMED（值域、預設值、寫入位址與延遲倍率）**  
日期：2026-08-02

## 證據身分

- 輸入：`WAR.EXE`
- SHA-256：`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`
- 工具：IDA Pro 9.4，Docker image `ida-pro-9.4-ver3`
- 位址空間：IDA linear address
- 非破壞匯出：
  - `workplace/ida/user-output/function-sub_1C38C.txt`
  - `workplace/ida/user-output/function-sub_1C220.txt`
  - `workplace/ida/user-output/xref-byte_6FE85.txt`
  - `workplace/ida/user-output/function-sub_20447.txt`

匯出均從唯讀 `WAR.EXE` 在 `/tmp` 重建一次性 `.i64`，沒有對 root-owned
`workplace/ida/WAR.EXE.i64` 執行 IDC，也沒有改名或寫入推測註解。

## 已證實結論

### 1. 其他選項第 6 項呼叫 `sub_1C220`

`sub_1C38C` 於 `1C51Ah..1C52Bh` 比對選項 6，呼叫 `sub_1C220`。

### 2. 值域是 1..10，寫入 `byte_6FE85`

`sub_1C220` 的輸入邊界：

- `1C2FAh` 拒絕小於 1。
- `1C308h` 拒絕大於 `0Ah`。
- `1C322h..1C325h` 將輸入的低位元組寫入 `byte_6FE85`。

`sub_391E1` 在 `39409h` 寫入 5，所以正常開局預設值是 5。

### 3. 每級是 400 毫秒

`byte_6FE85` 在 IDA 交叉參照圖中有 156 個直接讀取點。以
`sub_20447` 為診斷樣本，`204CCh..204D7h`、`20552h..2055Dh`、
`2077Dh..20788h` 及其他訊息分支都執行：

```text
mov al, byte_6FE85
xor ah, ah
mov dx, 190h
mul dx
push ax
call @DELAY$q4WORD
```

Turbo Pascal `DELAY` 的參數是毫秒，`0x190 = 400`，所以：

```text
停留時間 = byte_6FE85 × 400 ms
值域       = 400 ms .. 4,000 ms
預設       = 5 × 400 ms = 2,000 ms
```

此為「已證實」，不是從畫面數字猜測秒數。

## remake 對應

Ebiten 目前沒有改寫預設的每秒 60 次 `Update`，因此每級對應 24 ticks。
`messageQueue` 在訊息存在時繼續繪圖，但擋住新的玩家輸入；這對應原版
`DELAY` 期間不接受下一個指令的行為。多則訊息依發生順序排隊，不互相覆蓋。

設定屬裝置偏好，寫入 XDG `prefs.json` 的 `message_time`，不猜測或佔用
`.DT1` 未解位元組。
