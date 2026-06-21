# 微信小账本

轻量 Windows 个人账本，导入微信支付 Excel 流水，SQLite 本地存储，支持搜索与汇总。

## 功能

- 导入微信支付账单 Excel（表头含「交易单号」的标准导出格式）
- 以 **交易单号** 去重，重复导入自动跳过
- 流水列表：关键词、日期范围、收支、类型筛选
- 汇总：总收入/支出/结余、按类型、按月统计
- 数据默认保存在 `%AppData%\wxtrans\ledger.db`

## 构建

需要 Go 1.21+ 与 C 编译器（Fyne 在 Windows 上需要 CGO，通常安装 [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) 或 MinGW-w64）。

```powershell
go build -ldflags="-s -w -H windowsgui" -o wxtrans.exe .
```

或使用脚本：

```powershell
.\build.ps1
```

## 使用

1. 微信 → 我 → 服务 → 钱包 → 账单 → 右上角 … → 导出账单 → 选 Excel
2. 运行 `wxtrans.exe`，输入密码登录
3. 初始密码：`eee333`，登录后可在工具栏点击「改密码」修改
4. 点击「导入 Excel」，在「流水」页搜索筛选，在「汇总」页查看统计

## 快捷脚本

| 脚本 | 说明 |
|------|------|
| `build.bat` | 打包生成 `wxtrans.exe` |

## 命令行

```text
wxtrans.exe              # 启动 GUI
wxtrans.exe -db D:\path\ledger.db   # 指定数据库路径
```

## 示例文件

目录下的 `微信支付账单流水文件*.xlsx` 为官方导出样例，可直接用于测试导入。
