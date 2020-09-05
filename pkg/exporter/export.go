package exporter

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

func Export(filename string, marshaler func(interface{}) ([]byte, error)) (ret []byte, err error) {
	db, err := bolt.Open(filename, 0600, &bolt.Options{
		ReadOnly: true,
	})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return export(db, marshaler)
}

func export(db *bolt.DB, marshaler func(interface{}) ([]byte, error)) (ret []byte, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		rawMap := makeRawMap(tx, c)
		ret, err = marshaler(rawMap)
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func makeRawMap(tx *bolt.Tx, c *bolt.Cursor) map[string]interface{} {
	rawMap := make(map[string]interface{})
	recursiveRawMap(tx, c, rawMap)
	return rawMap
}

func recursiveRawMap(tx *bolt.Tx, c *bolt.Cursor, rawMap map[string]interface{}) map[string]interface{} {
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			bucket := c.Bucket().Bucket(k)
			if bucket == nil {
				bucket = tx.Bucket(k)
			}
			nextCursor := bucket.Cursor()
			nextMap := make(map[string]interface{})
			rawMap[string(k)] = recursiveRawMap(tx, nextCursor, nextMap)
			continue
		}

		// check bolt db value is json string or not.
		var rawJSON json.RawMessage
		if err := json.Unmarshal(v, &rawJSON); err != nil {
			// if the value is not json string, treat the the value as string
			rawMap[string(k)] = string(v)
		} else {
			rawMap[string(k)] = rawJSON
		}
	}
	return rawMap
}
