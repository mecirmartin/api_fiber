package database

import (
	"github.com/mecirmartin/fiber_api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := "host=localhost user=martinmecir password=root dbname=fiber_api port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	DB = connection

	if err != nil {
		panic("Could not connect to DB :(")
	}

	DB.AutoMigrate(models.User{})

}
