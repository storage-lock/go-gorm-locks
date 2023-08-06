package gorm_locks

import "gorm.io/gorm"

var GlobalGormLockFactory *GormLockFactory

func InitGormLockFactory(db *gorm.DB) error {
	factory, err := NewGormLockFactory(db)
	if err != nil {
		return err
	}
	GlobalGormLockFactory = factory
	return nil
}
