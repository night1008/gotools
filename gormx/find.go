package gormx

import (
	"errors"

	"gorm.io/gorm"
)

func FindOne(db *gorm.DB, out interface{}) (bool, error) {
	result := db.First(out)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// 不打印 record not found 日志
func FindOneIgnoreLog(db *gorm.DB, out interface{}) (bool, error) {
	result := db.Limit(1).Find(out)
	if err := result.Error; err != nil {
		return false, err
	}
	return result.RowsAffected == 1, nil
}

func FindOneWithNotFoundError(db *gorm.DB, out interface{}, notFoundErr error) error {
	if exist, err := FindOne(db, out); err != nil {
		return err
	} else if !exist {
		return notFoundErr
	}
	return nil
}
