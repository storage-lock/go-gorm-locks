package gorm_locks

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// sqlite3Dialector 让 db.Name() 返回 sqldb 认识的驱动名，走通第一条路。
type sqlite3Dialector struct {
	gorm.Dialector
}

func (sqlite3Dialector) Name() string { return "sqlite3" }

func TestCreateStorageForGormDb_FirstPathSuccess(t *testing.T) {
	// 真 sqlite 内存库同时做 ConnPool：第一条路 Name()=sqlite3 走通，
	// Init 的建表直接在真库上执行；Take 经 db.DB() 拿到同一真库。
	realDB, err := sql.Open("sqlite3", "file:gorm_firstpath.db?mode=memory")
	assert.Nil(t, err)
	defer realDB.Close()

	db := &gorm.DB{Config: &gorm.Config{Dialector: sqlite3Dialector{}, ConnPool: realDB}}
	m := NewGormConnectionManager(db)

	s, err := CreateStorageForGormDb(db, m)
	assert.Nil(t, err)
	assert.NotNil(t, s)

	factory, err := NewGormLockFactory(db)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	// 端到端加解锁
	lock, err := factory.CreateLock("test-gorm-firstpath-lock")
	assert.Nil(t, err)
	assert.NotNil(t, lock)
	assert.Nil(t, lock.Lock(context.Background(), "owner-1"))
	assert.Nil(t, lock.UnLock(context.Background(), "owner-1"))
}

type bogusDialector struct {
	gorm.Dialector
}

func (bogusDialector) Name() string { return "bogus-driver" }

func TestCreateStorageForGormDb_DBError(t *testing.T) {
	// Dialector 非法（第一条路失败）+ ConnPool 为 nil（db.DB() 返回 ErrInvalidDB）
	db := &gorm.DB{Config: &gorm.Config{Dialector: bogusDialector{}}}
	m := NewGormConnectionManager(db)

	_, err := CreateStorageForGormDb(db, m)
	assert.NotNil(t, err)

	// Take 的错误分支
	_, err = m.Take(t.Context())
	assert.NotNil(t, err)

	// NewGormLockFactory / InitGormLockFactory 的错误分支
	_, err = NewGormLockFactory(db)
	assert.NotNil(t, err)

	GlobalGormLockFactory = nil
	assert.NotNil(t, InitGormLockFactory(db))
	assert.Nil(t, GlobalGormLockFactory)
}

func TestCreateStorageForGormDb_UnsupportedSqlDriver(t *testing.T) {
	// ConnPool 是 sqlmock 的 *sql.DB：db.DB() 成功，但 NewStorageBySqlDb 识别驱动失败
	mockDB, _, err := sqlmock.New()
	assert.Nil(t, err)
	defer mockDB.Close()

	db := &gorm.DB{Config: &gorm.Config{Dialector: sqlite.Dialector{}, ConnPool: mockDB}}
	m := NewGormConnectionManager(db)

	// sqlite 驱动的 Name() 返回 "sqlite"，与 sqldb 的 "sqlite3" 常量对不上，
	// 第一条路失败后回退到 db.DB()（成功），再因驱动不支持而失败
	_, err = CreateStorageForGormDb(db, m)
	assert.NotNil(t, err)

	_, err = NewGormLockFactory(db)
	assert.NotNil(t, err)
}
