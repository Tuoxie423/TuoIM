package initial

import (
	"fmt"
	"ginchat/config"
	"ginchat/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitMySQL() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Global.MySQL.User,
		config.Global.MySQL.Password,
		config.Global.MySQL.Host,
		config.Global.MySQL.Port,
		config.Global.MySQL.Name)
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		log.Fatal("Faild to connent Mysql:", err)
	}
	db.AutoMigrate(
		&models.UserBasic{},
	)
	config.Global.DB = db
}
