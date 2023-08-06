package gorm_locks

import (
	"database/sql"
	sqldb_storage "github.com/storage-lock/go-sqldb-storage"
	"github.com/storage-lock/go-storage"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
	"gorm.io/gorm"
)

type GormLockFactory struct {
	db *gorm.DB
	*storage_lock_factory.StorageLockFactory[*sql.DB]
}

func NewGormLockFactory(db *gorm.DB) (*GormLockFactory, error) {
	connectionManager := NewGormConnectionManager(db)

	storage, err := CreateStorageForGormDb(db, connectionManager)
	if err != nil {
		return nil, err
	}

	factory := storage_lock_factory.NewStorageLockFactory[*sql.DB](storage, connectionManager)

	return &GormLockFactory{
		db:                 db,
		StorageLockFactory: factory,
	}, nil
}

// CreateStorageForGormDb 尝试从GORM创建Storage
func CreateStorageForGormDb(db *gorm.DB, connectionManager storage.ConnectionManager[*sql.DB]) (storage.Storage, error) {

	// 先尝试根据驱动名称创建
	storage, err := sqldb_storage.NewStorageByDriverName(db.Name(), connectionManager)
	if storage != nil && err == nil {
		return storage, err
	}

	// 再然后根据识别出来的名称创建
	// TODO 2023-8-6 23:13:53 确认是否有连接泄露风险
	s, err := db.DB()
	if err != nil {
		return nil, err
	}
	return sqldb_storage.NewStorageBySqlDb(s, connectionManager)
}
