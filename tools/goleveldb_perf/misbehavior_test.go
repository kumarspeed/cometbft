package db_experiments

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/rand"
	"github.com/docker/go-units"
)

func TestBoltDBInt32Overflow(t *testing.T) {
	config := test.ResetTestRoot("db_benchmark")
	defer func() {
		err := os.RemoveAll(config.RootDir)
		if err != nil {
			panic(err)
		}
	}()

	db, err := dbm.NewDB("experiment_db", dbm.BoltDBBackend, config.DBDir())
	if err != nil {
		panic(err)
	}

	// For some reason, due to an iterator failure,
	// this defer call hangs forever
	//
	//defer func(db dbm.DB) {
	//	err := db.Close()
	//	if err != nil {
	//		panic(err)
	//	}
	//}(db)

	for i := 0; i < 4; i++ {
		err := db.Set(rand.Bytes(64), rand.Bytes(512*units.MiB))
		if err != nil {
			panic(err)
		}

		iter, err := db.Iterator(nil, nil)
		if err != nil {
			panic(err)
		}

		for ; iter.Valid(); iter.Next() {
		}

		err = iter.Close()
		if err != nil {
			panic(err)
		}
	}
}

func TestRocksDBWithIterator(t *testing.T) {
	config := test.ResetTestRoot("db_benchmark")
	defer func() {
		err := os.RemoveAll(config.RootDir)
		if err != nil {
			panic(err)
		}
	}()

	db, err := dbm.NewDB("experiment_db", dbm.RocksDBBackend, config.DBDir())
	if err != nil {
		panic(err)
	}
	defer func(db dbm.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	for i := 0; i < 10; i++ {
		batch := db.NewBatch()
		for j := 0; j < 1000; j++ {
			err := batch.Set(rand.Bytes(64), rand.Bytes(1*units.MiB))
			if err != nil {
				panic(err)
			}
		}
		err := batch.WriteSync()
		if err != nil {
			panic(err)
		}
		err = batch.Close()
		if err != nil {
			panic(err)
		}

		iter, err := db.Iterator(nil, nil)
		if err != nil {
			panic(err)
		}
		for ; iter.Valid(); iter.Next() {
		}
		err = iter.Close()
		if err != nil {
			panic(err)
		}
	}
}

func TestRocksDBWithoutIterator(t *testing.T) {
	config := test.ResetTestRoot("db_benchmark")
	defer func() {
		err := os.RemoveAll(config.RootDir)
		if err != nil {
			panic(err)
		}
	}()

	db, err := dbm.NewDB("experiment_db", dbm.RocksDBBackend, config.DBDir())
	if err != nil {
		panic(err)
	}
	defer func(db dbm.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	for i := 0; i < 10; i++ {
		batch := db.NewBatch()
		for j := 0; j < 1000; j++ {
			err := batch.Set(rand.Bytes(64), rand.Bytes(1*units.MiB))
			if err != nil {
				panic(err)
			}
		}
		err := batch.WriteSync()
		if err != nil {
			panic(err)
		}
		err = batch.Close()
		if err != nil {
			panic(err)
		}
	}
}

func TestBoltDBDeletions(t *testing.T) {
	config := test.ResetTestRoot("db_benchmark")
	defer func() {
		err := os.RemoveAll(config.RootDir)
		if err != nil {
			panic(err)
		}
	}()

	db, err := dbm.NewDB("experiment_db", dbm.BoltDBBackend, config.DBDir())
	if err != nil {
		panic(err)
	}

	// This call never returns preventing test from completing
	//
	//defer func(db dbm.DB) {
	//	err := db.Close()
	//	if err != nil {
	//		panic(err)
	//	}
	//}(db)

	nBulks := 7
	for j := 0; j < nBulks; j++ {
		fmt.Println(fmt.Sprintf("Inserting %v out of %v", j+1, nBulks))
		for i := 0; i < 10; i++ {
			err = db.Set(rand.Bytes(64), rand.Bytes(100*units.MiB))
			if err != nil {
				panic(err)
			}
		}
	}

	runInTime := runWithTimeOut(10*time.Minute, func(ctx context.Context) {
		deleted := 0
		iter, err := db.Iterator(nil, nil)
		if err != nil {
			panic(err)
		}
		for ; iter.Valid(); iter.Next() {
			err = db.Delete(iter.Key())
			if err != nil {
				panic(err)
			}
			deleted++
			fmt.Println("Deleted ", deleted)
		}
		func(iter dbm.Iterator) {
			err := iter.Close()
			if err != nil {
				panic(err)
			}
		}(iter)
	})

	if !runInTime {
		t.Log("Timed out")
		t.Fail()
	}
}