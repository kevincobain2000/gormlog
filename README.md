<h1 align="center">
  Missing Gorm log/slog
</h1>

Use GO's `log/slog` as Gorm logger.

## Usage

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    gormlog "github.com/kevincobain2000/gormlog/slog"
)

dsn := "user:pass@tcp(127.0.0.1:3306)/mydb"
db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: &gormlog.Slog{}, // will hook up slog to gorm's logger
})
```