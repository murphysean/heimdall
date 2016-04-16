package sqldb

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
)

func genUUIDv4() string {
	u := make([]byte, 16)
	rand.Read(u)
	//Set the version to 4
	u[6] = (u[6] | 0x40) & 0x4F
	u[8] = (u[8] | 0x80) & 0xBF
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

type SqlDB struct {
	Db *sql.DB
}

func NewSqlDB(db *sql.DB) *SqlDB {
	sdb := new(SqlDB)
	sdb.Db = db
	db.Begin()

	check(db.Exec("CREATE TABLE IF NOT EXISTS clients (id TEXT NOT NULL PRIMARY KEY, name TEXT NOT NULL, secret TEXT NOT NULL, type TEXT NOT NULL, internal INTEGER NOT NULL DEFAULT 0, redirecturis TEXT NOT NULL)"))
	check(db.Exec("CREATE TABLE IF NOT EXISTS users (id TEXT NOT NULL PRIMARY KEY, name TEXT NOT NULL, json TEXT)"))
	check(db.Exec("CREATE TABLE IF NOT EXISTS auth (userid TEXT NOT NULL PRIMARY KEY, username TEXT NOT NULL UNIQUE, password TEXT NOT NULL, salt TEXT NOT NULL, FOREIGN KEY userid REFERENCES users(id) ON DELETE CASCADE"))
	check(db.Exec("CREATE TABLE IF NOT EXISTS tokens (id TEXT NOT NULL PRIMARY KEY, type TEXT NOT NULL, userid TEXT NOT NULL, clientid TEXT NOT NULL, expires DATETIME NOT NULL, scope TEXT NOT NULL, accesstype TEXT NOT NULL, refreshtokenid TEXT NOT NULL, FOREIGN KEY userid REFERENCES users(id) ON DELETE CASCADE, FOREIGN KEY clientid REFERENCES clients(id) ON DELETE CASCADE, FOREIGN KEY refreshtokenid REFERENCES tokens(id) ON DELETE CASCADE)"))
	check(db.Exec("CREATE TABLE IF NOT EXISTS concents (userid TEXT NOT NULL, clientid TEXT NOT NULL, concent TEXT NOT NULL, PRIMARY KEY(userid,clientid,concent), FOREIGN KEY userid REFERENCES users(id) ON DELETE CASCADE, FOREIGN KEY clientid REFERENCES clients(id) ON DELETE CASCADE)"))

	return sdb
}

func check(r sql.Result, err error) {
	if err != nil {
		log.Fatal(err)
	}
}
