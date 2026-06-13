# Design: queryli — Multi-Database CLI Query Tool

---

## 1. Mô tả hệ thống

**queryli** là CLI tool viết bằng Go, cho phép người dùng:
- Quản lý nhiều **connection profiles** (mỗi profile = thông tin đăng nhập + loại database)
- **Connect** tới một database, duy trì connection sống dưới dạng **daemon process**
- Thực thi **SQL query** hoặc chạy **file .sql** thông qua connection đang mở
- Chỉ cho phép **1 connection active tại 1 thời điểm**
- Hỗ trợ: **PostgreSQL, MySQL, SQLite, Oracle** (dễ mở rộng thêm)

**Người dùng:** Developer, DBA, DevOps.

**Kiến trúc cốt lõi:** Daemon/Agent pattern — lệnh `connect` spawn một background process giữ DB connection, các lệnh tiếp theo giao tiếp với daemon qua Unix Domain Socket.

---

## 2. Tech Stack

| Thành phần | Lựa chọn | Lý do |
|---|---|---|
| Language | Go | Single binary, cross-platform, ecosystem mạnh cho CLI |
| CLI framework | `spf13/cobra` | De-facto standard (kubectl, docker, gh) |
| Config | `spf13/viper` | Đọc YAML, env var, tích hợp cobra |
| DB abstraction | `database/sql` (stdlib) | Interface chuẩn, scan dynamic columns, không cần ORM |
| PostgreSQL | `jackc/pgx/v5/stdlib` | Actively maintained, recommended |
| MySQL | `go-sql-driver/mysql` | Phổ biến nhất, ổn định |
| SQLite | `mattn/go-sqlite3` | Phổ biến nhất cho SQLite |
| Oracle | `sijms/go-ora/v2` | Pure Go, không cần Oracle Client |
| Output | `olekukonko/tablewriter` | In kết quả dạng bảng |
| IPC | Unix Domain Socket + JSON | Đơn giản, nhanh, không cần port |

---

## 3. Cấu trúc lệnh

```
queryli
│
├── profile add        Thêm connection profile
├── profile list       Liệt kê tất cả profiles
├── profile remove     Xóa một profile
├── profile use        Đặt profile làm active mặc định
│
├── connect [profile]  Tạo daemon, mở connection tới DB, giữ sống
├── disconnect         Tắt daemon, đóng connection
├── status             Xem daemon đang connect tới profile nào
│
├── query <sql>        Gửi câu SQL tới daemon, nhận kết quả
├── exec  <file.sql>   Gửi nội dung file .sql tới daemon
└── ping               Kiểm tra daemon + DB connection còn sống không
```

### Global Flags

```
--profile, -p   <name>           Override active profile (dùng cho connect)
--config        <path>           Đường dẫn config file tùy chỉnh
--password      <value>          Password (không lưu vào config)
--format, -f    table|json|csv   Định dạng output (default: table)
```

### Ví dụ sử dụng

```bash
# --- Quản lý profile ---
queryli profile add --name prod-pg --type postgres \
    --host db.prod.com --port 5432 --user admin --db myapp
queryli profile add --name local-sqlite --type sqlite --path ./dev.db
queryli profile add --name erp-oracle --type oracle \
    --host oracle.corp.internal --port 1521 --user app_user --service ORCL
queryli profile list
queryli profile use prod-pg

# --- Kết nối ---
queryli connect                        # connect tới active profile
queryli connect staging-mysql          # connect tới profile cụ thể
# → Daemon started. Connected to [staging-mysql] (mysql@staging.db:3306)

# --- Truy vấn (daemon đang chạy) ---
queryli query "SELECT * FROM users LIMIT 5"
queryli query "SELECT count(*) FROM orders" --format json
queryli exec ./scripts/seed.sql

# --- Kiểm tra ---
queryli ping
# → Connected to [staging-mysql] (mysql@staging.db:3306) — 2ms

queryli status
# → Daemon running (PID 12345)
# → Profile: staging-mysql
# → Type: mysql
# → Host: staging.db:3306
# → Uptime: 15m32s

# --- Ngắt kết nối ---
queryli disconnect
# → Disconnected from [staging-mysql]. Daemon stopped.

# --- Chuyển sang DB khác ---
queryli connect prod-pg --password
# → Enter password: ****
# → Daemon started. Connected to [prod-pg] (postgres@db.prod.com:5432)
```

---

## 4. Config File

Lưu tại `~/.queryli/config.yaml`:

```yaml
active_profile: prod-pg

profiles:
  prod-pg:
    type: postgres
    host: db.prod.com
    port: 5432
    user: admin
    password: ""
    dbname: myapp
    sslmode: require

  staging-mysql:
    type: mysql
    host: staging.db.internal
    port: 3306
    user: dev
    password: ""
    dbname: staging_db

  local-sqlite:
    type: sqlite
    path: ./dev.db

  erp-oracle:
    type: oracle
    host: oracle.corp.internal
    port: 1521
    user: app_user
    password: ""
    service: ORCL           # service name hoặc SID
```

### State files (runtime)

```
~/.queryli/
├── config.yaml       # profiles + active profile
├── daemon.pid        # PID của daemon đang chạy
├── daemon.sock       # Unix socket cho IPC
└── daemon.meta       # Metadata: profile name, connect time
```

> **Bảo mật:** Password không nên lưu trong config.yaml. Sử dụng flag `--password` (prompt nhập) hoặc env var `QUERYLI_PASSWORD`.

---

## 5. Kiến trúc tổng thể

```
                    ┌─────────────────────────────────────┐
                    │          ~/.queryli/                 │
                    │  config.yaml  daemon.pid  daemon.sock│
                    └──────┬──────────┬──────────┬────────┘
                           │          │          │
    ┌──────────┐     load config   check PID   connect
    │  User    │           │          │          │
    │ Terminal ├───┐       │          │          │
    └──────────┘   │       ▼          ▼          ▼
                   │  ┌──────────────────────────────┐
  queryli connect ─┼─▶│  Connect Command              │
                   │  │  1. Load profile from config   │
                   │  │  2. Check daemon.pid → chưa?   │
                   │  │  3. Spawn daemon process       │
                   │  │  4. Lưu PID + socket + meta    │
                   │  └──────────┬───────────────────┘
                   │             │ os/exec (background)
                   │             ▼
                   │  ┌──────────────────────────────┐
                   │  │  Daemon Process               │
                   │  │  ┌────────────────────────┐   │
                   │  │  │ db.Connector            │   │
                   │  │  │  • Connect()            │   │
                   │  │  │  • Query() / Exec()     │   │
                   │  │  │  • Ping() / Close()     │   │
                   │  │  └────────────────────────┘   │
                   │  │          ▲                     │
                   │  │          │ database/sql        │
                   │  │          ▼                     │
                   │  │  ┌────────────────────────┐   │
                   │  │  │ PostgreSQL / MySQL      │   │
                   │  │  │ / SQLite (actual DB)    │   │
                   │  │  └────────────────────────┘   │
                   │  │                               │
                   │  │  Unix Socket Server            │
                   │  │  daemon.sock ◄────────────┐   │
                   │  └──────────────────────────┼───┘
                   │                             │
  queryli query ───┼─▶ IPC Client ───────────────┘
  queryli exec  ───┤     gửi JSON request
  queryli ping  ───┤     nhận JSON response
  queryli disconnect──▶  gửi shutdown signal
```

---

## 6. IPC Protocol

Giao tiếp qua Unix socket bằng JSON, mỗi message kết thúc bằng `\n` (newline-delimited JSON).

### Request

```go
type Request struct {
    Type string `json:"type"`   // "query" | "exec" | "ping" | "status" | "shutdown"
    SQL  string `json:"sql"`    // câu SQL hoặc nội dung file
}
```

### Response

```go
type Response struct {
    OK           bool       `json:"ok"`
    Columns      []string   `json:"columns,omitempty"`
    Rows         [][]string `json:"rows,omitempty"`
    RowsAffected int64      `json:"rows_affected,omitempty"`
    Message      string     `json:"message,omitempty"`
    Error        string     `json:"error,omitempty"`
    Status       *Status    `json:"status,omitempty"`
}

type Status struct {
    Profile     string `json:"profile"`
    DBType      string `json:"db_type"`
    Host        string `json:"host"`
    Uptime      string `json:"uptime"`
}
```

### Ví dụ giao tiếp

```
→ {"type":"ping","sql":""}
← {"ok":true,"message":"pong","status":{"profile":"prod-pg","uptime":"5m12s"}}

→ {"type":"query","sql":"SELECT id, name FROM users LIMIT 2"}
← {"ok":true,"columns":["id","name"],"rows":[["1","Alice"],["2","Bob"]]}

→ {"type":"exec","sql":"INSERT INTO logs (msg) VALUES ('test')"}
← {"ok":true,"rows_affected":1,"message":"1 row(s) affected"}

→ {"type":"shutdown","sql":""}
← {"ok":true,"message":"bye"}
```

---

## 7. Core Interfaces

### 7.1 Connector — DB Abstraction

```go
// internal/db/connector.go

type Connector interface {
    Connect() error
    Query(sql string) (*QueryResult, error)
    Exec(sql string) (int64, error)
    Ping() error
    Close() error
}

type QueryResult struct {
    Columns []string
    Rows    [][]string
}

func NewConnector(profile config.Profile) (Connector, error) {
    switch profile.Type {
    case "postgres":
        return &PostgresConnector{profile: profile}, nil
    case "mysql":
        return &MySQLConnector{profile: profile}, nil
    case "sqlite":
        return &SQLiteConnector{profile: profile}, nil
    case "oracle":
        return &OracleConnector{profile: profile}, nil
    default:
        return nil, fmt.Errorf("unsupported database type: %s", profile.Type)
    }
}
```

Mỗi implementation bên dưới dùng `database/sql` + driver tương ứng. Query result dùng dynamic column scan:

```go
// Pseudo-code cho Query — dùng chung cho mọi driver
func (c *BaseConnector) query(db *sql.DB, rawSQL string) (*QueryResult, error) {
    rows, err := db.Query(rawSQL)
    cols, _ := rows.Columns()
    // scan từng row vào []interface{} → convert sang []string
    // return &QueryResult{Columns: cols, Rows: allRows}
}
```

### 7.2 Formatter — Output Abstraction

```go
// internal/output/formatter.go

type Formatter interface {
    Format(result *db.QueryResult) string
}

func NewFormatter(format string) Formatter {
    switch format {
    case "json":  return &JSONFormatter{}
    case "csv":   return &CSVFormatter{}
    default:      return &TableFormatter{}
    }
}
```

---

## 8. Package Structure

```
queryli/
├── main.go                         # cobra.Execute()
├── go.mod
├── go.sum
│
├── cmd/
│   ├── root.go                     # root command, global flags, viper init
│   ├── connect.go                  # spawn daemon, lưu PID
│   ├── disconnect.go               # gửi shutdown tới daemon
│   ├── status.go                   # hiển thị trạng thái daemon
│   ├── query.go                    # gửi query qua socket
│   ├── exec.go                     # đọc file .sql → gửi qua socket
│   ├── ping.go                     # kiểm tra daemon + DB
│   ├── profile.go                  # parent command cho profile group
│   ├── profile_add.go
│   ├── profile_list.go
│   ├── profile_remove.go
│   └── profile_use.go
│
├── internal/
│   ├── config/
│   │   └── config.go               # Config, Profile struct; Load/Save
│   │
│   ├── daemon/
│   │   ├── daemon.go               # Daemon struct: giữ Connector + Socket server
│   │   ├── handler.go              # Xử lý từng request type
│   │   ├── lifecycle.go            # Start / Stop / PID file management
│   │   └── meta.go                 # Đọc/ghi daemon.meta (profile, start time)
│   │
│   ├── db/
│   │   ├── connector.go            # Connector interface + factory
│   │   ├── base.go                 # Base implementation (shared query logic)
│   │   ├── postgres.go
│   │   ├── mysql.go
│   │   ├── sqlite.go
│   │   └── oracle.go
│   │
│   ├── ipc/
│   │   ├── protocol.go             # Request, Response, Status struct
│   │   ├── client.go               # Dial socket → send request → read response
│   │   └── server.go               # Listen socket → accept → dispatch handler
│   │
│   └── output/
│       ├── formatter.go            # Formatter interface + factory
│       ├── table.go
│       ├── json.go
│       └── csv.go
│
└── README.md
```

---

## 9. Daemon Lifecycle

### Start (queryli connect)

```
1. Load config → resolve profile (từ arg hoặc active_profile)
2. Kiểm tra ~/.queryli/daemon.pid
   ├── PID tồn tại + process sống → báo lỗi: "already connected to [X], run disconnect first"
   └── Không có hoặc process đã chết → tiếp tục
3. Spawn daemon process:
   exec.Command(os.Args[0], "__daemon", "--profile", profileName)
   └── Detach: cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
   └── Redirect stdout/stderr → ~/.queryli/daemon.log
4. Daemon process:
   a. Ghi PID → daemon.pid
   b. Ghi metadata → daemon.meta
   c. db.NewConnector(profile) → connector.Connect()
   d. ipc.NewServer("daemon.sock") → server.Listen()
   e. Vòng lặp: accept connection → đọc request → handler → gửi response
5. Client nhận PID, verify daemon sống → in success message
```

> `__daemon` là một hidden subcommand, user không thấy trong help. Client gọi chính binary của mình với subcommand này để spawn daemon.

### Stop (queryli disconnect)

```
1. Đọc daemon.pid → kiểm tra process sống
2. Gửi {"type":"shutdown"} qua socket
3. Daemon nhận → connector.Close() → server.Close() → os.Exit(0)
4. Client xóa daemon.pid, daemon.sock, daemon.meta
```

### Crash Recovery

```
queryli connect / query / ping:
  → Đọc daemon.pid
  → os.FindProcess(pid) → process.Signal(0)
  → Nếu process chết: xóa stale PID + socket, báo "not connected"
```

---

## 10. Entity: Config & Profile

```go
type Config struct {
    ActiveProfile string             `yaml:"active_profile"`
    Profiles      map[string]Profile `yaml:"profiles"`
}

type Profile struct {
    Type     string `yaml:"type"`       // postgres | mysql | sqlite | oracle
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    DBName   string `yaml:"dbname"`
    SSLMode  string `yaml:"sslmode"`    // postgres
    Path     string `yaml:"path"`       // sqlite
    Service  string `yaml:"service"`    // oracle (service name hoặc SID)
}
```

---

## 11. Todo-list: Thứ tự triển khai

### 1. Project setup + Config system
- Scope: init Go module, cài cobra/viper, struct Config/Profile, Load/Save YAML, tạo `~/.queryli/`
- Lý do làm trước: mọi feature đều phụ thuộc config

### 2. Profile management
- Scope: `profile add`, `profile list`, `profile remove`, `profile use`
- Không cần kết nối DB thật
- Phụ thuộc: Feature 1

### 3. IPC protocol + Socket server/client
- Scope: định nghĩa Request/Response struct, implement Unix socket server + client, newline-delimited JSON
- Lý do: daemon và các command đều giao tiếp qua layer này
- Phụ thuộc: Feature 1

### 4. Connector Interface + Implementations *(song song với Feature 3)*
- Scope: Connector interface, factory, PostgresConnector, MySQLConnector, SQLiteConnector, OracleConnector, dynamic column scan
- Phụ thuộc: Feature 1

### 5. Daemon core
- Scope: daemon process, handler, PID file management, meta file, hidden `__daemon` subcommand, crash recovery
- Phụ thuộc: Feature 3, 4

### 6. `connect` + `disconnect` + `status` + `ping`
- Scope: spawn daemon, shutdown daemon, kiểm tra trạng thái
- Phụ thuộc: Feature 5

### 7. `query` + `exec` + Output Formatter
- Scope: gửi SQL qua socket, đọc file .sql, format kết quả (table/json/csv)
- Phụ thuộc: Feature 5, 6
- Đây là feature chính, nhưng tới đây mọi layer bên dưới đã sẵn sàng

### 8. Polish
- Scope: password prompt, env var `QUERYLI_PASSWORD`, mask password trong `profile list`, error messages, daemon.log
- Phụ thuộc: tất cả feature trên
- Không ảnh hưởng logic chính, làm cuối

---

## 12. Mở rộng tương lai

- **MSSQL support** — thêm 1 file connector + 1 case trong factory
- **Interactive REPL mode** — `queryli shell` mở readline loop, gửi qua socket
- **Query history** — lưu lịch sử query vào `~/.queryli/history`
- **Export** — `queryli query "..." --format csv > output.csv`
- **Connection timeout** — daemon tự tắt sau N phút idle
- **TLS/SSH tunnel** — hỗ trợ kết nối qua SSH tunnel tới remote DB
