package database

import (
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

// Batch run the leveldb batch
type Batch struct {
	connstr string
}

func OpenDB(dbname string) *leveldb.DB {
	db, err := leveldb.OpenFile("./data/"+dbname, nil)
	if err != nil {
		log.Fatal("open error")
	}
	return db
}

func Put(db *leveldb.DB, key []byte, value []byte) {
	if err := db.Put(key, value, nil); err != nil {
		log.Fatal("put error")
	}
}

func Del(db *leveldb.DB, key []byte) {
	if err := db.Delete(key, nil); err != nil {
		log.Fatal("delete error")
	}
}

// Exist check the key exists
func Exist(db *leveldb.DB, key string) bool {
	if _, err := db.Get([]byte(key), nil); err != nil {
		return false
	} else {
		return true
	}
}
