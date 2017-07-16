package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/revel/revel"
)

var DB *sql.DB

func InitDB() {
	dbUser := revel.Config.StringDefault("db.user", "root")
	dbPassword := revel.Config.StringDefault("db.password", "")
	dbAddr := revel.Config.StringDefault("db.addr", "127.0.0.1:3306")
	dbName := revel.Config.StringDefault("db.name", "")

	var err error
	DB, err = sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/%s",
			dbUser, dbPassword, dbAddr, dbName))
	if err != nil {
		revel.ERROR.Fatal(err)
	}
}
