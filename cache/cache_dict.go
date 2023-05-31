package cache

import (
	"r3/db"
	"r3/tools"
	"sync"
)

var (
	dict    []string // list of dictionaries for full text search, read from DB
	dict_mx sync.Mutex
)

func GetSearchDictionaries() []string {
	dict_mx.Lock()
	defer dict_mx.Unlock()
	return dict
}

func GetSearchDictionaryIsValid(entry string) bool {
	dict_mx.Lock()
	defer dict_mx.Unlock()
	return tools.StringInSlice(entry, dict)
}

func LoadSearchDictionaries() error {
	dict_mx.Lock()
	defer dict_mx.Unlock()

	err := db.Pool.QueryRow(db.Ctx, `
		SELECT ARRAY_AGG(cfgname::TEXT)
		FROM pg_catalog.pg_ts_config
	`).Scan(&dict)

	return err
}