package main

import (
    "database/sql"
    "log"
    // _ "github.com/go-sql-driver/mysql" // Removed as we are switching to SQLite and GORM
)

var db *sql.DB

func initDB() {
    var err error
    dsn := "test"
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal("error:", err)
    }
}

func main() {
    log.Println("test")
}
