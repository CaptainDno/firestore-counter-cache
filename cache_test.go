package go_counter_map

import (
	"context"
	"errors"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const entryCount = 5
const fieldName = "test_counter"

func docId(k int) string {
	return fmt.Sprintf("test/%d", k)
}

func count(cache *CounterCache[int, uint32], k int, v uint32, wg *sync.WaitGroup) {
	var i uint32
	for i = 0; i < v; i++ {
		cache.Update(k, 1)
	}
	wg.Done()
}

func Test(t *testing.T) {
	err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
	t.Logf("Starting test of the counter cache")
	// Init firebase
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "nuclearquests"}
	app, err := firebase.NewApp(ctx, conf, option.WithCredentialsFile("G:\\nuclearquests-firebase-adminsdk-6wlqo-bfca4551c1.json"))
	if err != nil {
		t.Fatal(err)
	}
	db, err := app.Firestore(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := make(map[int]uint32, entryCount)
	for i := 0; i < entryCount; i++ {
		key := rand.Int()
		value := rand.Uint32()
		t.Logf("Generated key %d for value %d", key, value)
		_, err = db.Collection("test").Doc(strconv.Itoa(key)).Set(ctx, map[string]interface{}{
			fieldName: 0,
		})
		if err != nil {
			t.Fatal(err)
		}
		m[key] = value
	}
	cache := NewUInt32Cache[int](db, docId, fieldName)
	var wg sync.WaitGroup
	first := true
	for k, v := range m {
		if first {
			t.Log("Entry level testing")
			cache.Update(k, 1)
			if !(len(*cache.readMap.Load()) == 0 && len(cache.dirtyMap) == 1) {
				t.Fatal("The entry was not created in dirty map as expected")
			}
			cache.Update(k, 0)
			cache.Update(k, 0)
			if !(len(*cache.readMap.Load()) == 1 && len(cache.dirtyMap) == 0) {
				t.Fatal("The entry was not moved to fast read only map")
			}
			cache.Update(k, v-1)
			cache.Commit(ctx)
			cache.Commit(ctx)
			if !(len(*cache.readMap.Load()) == 0 && len(cache.dirtyMap) == 1) {
				t.Fatal("The entry was not moved to dirty map as expected")
			}
			cache.Commit(ctx)
			if !(len(*cache.readMap.Load()) == 0 && len(cache.dirtyMap) == 0) {
				t.Fatal("The entry was not deleted from dirty map as expected")
			}
			t.Log("OK")
			first = false
			continue
		}
		wg.Add(1)
		go count(cache, k, v, &wg)
	}
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			t.Log("Committing changes")
			cache.Commit(ctx)
			t.Logf("Fast cache size len: %d\tDirty cache len: %d", len(*cache.readMap.Load()), len(cache.dirtyMap))
		}
	}()
	wg.Wait()
	cache.Commit(ctx)
	iter := db.Collection("test").Documents(ctx)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data := doc.Data()
		id, err := strconv.Atoi(doc.Ref.ID)
		if err != nil {
			t.Fatal(err)
		}
		value := data[fieldName]
		expected := m[id]
		if value != expected {
			t.Errorf("Unexpected value in DB: %d instead of %d", value, expected)
			t.Fail()
		}
	}
}
