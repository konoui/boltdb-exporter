package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/google/go-cmp/cmp"
)

const filename = "boltdb-test.db"

type TestData struct {
	Data1 string
	Data2 string
	Data3 int
	Data4 *TestData
	Data5 *TestData
}

var testData = TestData{
	Data1: "1",
	Data2: "2",
	Data3: 3,
	Data4: &TestData{
		Data1: "4-1",
		Data2: "4-2",
		Data3: 3,
	},
	Data5: &TestData{
		Data1: "data5-nested1",
		Data2: "data5-nested2",
		Data4: &TestData{
			Data1: "5-4-1",
			Data2: "5-4-2",
			Data3: 541,
		},
	},
}

func Test_export(t *testing.T) {
	tests := []struct {
		name      string
		wantPath  string
		setuper   func(*testing.T, string) *bolt.DB
		update    bool
		selection map[string]bool
	}{
		{
			name:      "multi buckets",
			wantPath:  filepath.Join("testdata", "single-output.json"),
			setuper:   setupSingle,
			selection: nil,
		},
		{
			name:      "nested buckets",
			setuper:   setupNested,
			wantPath:  filepath.Join("testdata", "nested-output.json"),
			selection: nil,
		},
		{
			name:      "bucket selection",
			wantPath:  filepath.Join("testdata", "single-output-selection.json"),
			setuper:   setupSingle,
			selection: map[string]bool{"bucket2": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setuper(t, filename)
			defer cleanup(t, db, filename)

			gotData, err := export(db, json.Marshal, tt.selection)
			if err != nil {
				t.Error(err)
			}
			if tt.update {
				if err := os.WriteFile(tt.wantPath, gotData, 0644); err != nil {
					t.Fatal(err)
				}
			}

			wantData, err := os.ReadFile(tt.wantPath)
			if err != nil {
				t.Fatal(err)
			}
			if diff := diffJSON(wantData, gotData); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func setupSingle(t *testing.T, filename string) *bolt.DB {
	db := setupDB(t, filename)
	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}
	if err := singleData(db, "bucket1", data); err != nil {
		t.Fatal(err)
	}
	if err := singleData(db, "bucket1", data); err != nil {
		t.Fatal(err)
	}
	if err := singleData(db, "bucket1", data); err != nil {
		t.Fatal(err)
	}
	if err := singleData(db, "bucket2", data); err != nil {
		t.Fatal(err)
	}
	if err := singleData(db, "bucket2", data); err != nil {
		t.Fatal(err)
	}
	if err := singleData(db, "bucket3", data); err != nil {
		t.Fatal(err)
	}

	return db
}

func setupNested(t *testing.T, filename string) *bolt.DB {
	db := setupDB(t, filename)
	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}
	if err := nestedData(db, "bucketA", data); err != nil {
		t.Fatal(err)
	}
	if err := nestedData(db, "bucketB", data); err != nil {
		t.Fatal(err)
	}
	if err := nestedData(db, "bucketB", data); err != nil {
		t.Fatal(err)
	}

	return db
}

func setupDB(t *testing.T, filename string) *bolt.DB {
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func singleData(db *bolt.DB, bucket string, data []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		id, _ := b.NextSequence()
		return b.Put(itob(id), data)
	})
}

func nestedData(db *bolt.DB, bucket string, data []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		nested, err := b.CreateBucketIfNotExists([]byte(fmt.Sprintf("%s-nested", bucket)))
		if err != nil {
			return err
		}
		id, _ := nested.NextSequence()
		err = nested.Put(itob(id), data)
		if err != nil {
			return err
		}

		nestedNested, err := nested.CreateBucketIfNotExists([]byte(fmt.Sprintf("%s-nested-nested", bucket)))
		if err != nil {
			return err
		}

		id, _ = nestedNested.NextSequence()
		return nestedNested.Put(itob(id), data)
	})
}

func cleanup(t *testing.T, db *bolt.DB, filename string) {
	db.Close()
	if err := os.Remove(filename); err != nil {
		t.Fatal(err)
	}
}

func itob(v uint64) []byte {
	str := strconv.FormatUint(v, 10)
	return []byte(str)
}

func diffJSON(wantData, gotData []byte) string {
	want := make(map[string]interface{})
	got := make(map[string]interface{})

	if err := json.Unmarshal(wantData, &want); err != nil {
		return fmt.Sprintf("Unmarshal Error in wantData: %+v\n, string(wantData): %s\n, string(gotData): %s\n", err, string(wantData), string(gotData))
	}

	if err := json.Unmarshal(gotData, &got); err != nil {
		return fmt.Sprintf("Unmarshal Error in gotData: %+v\n, string(wantData): %s\n, string(gotData): %s\n", err, string(wantData), string(gotData))
	}

	return cmp.Diff(want, got)
}
