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
	serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4", config.Global.MySQL.User,
		config.Global.MySQL.Password,
		config.Global.MySQL.Host,
		config.Global.MySQL.Port)
	serverDB, _ := gorm.Open(mysql.Open(serverDSN))
	serverDB.Exec("CREATE DATABASE IF NOT EXISTS `ginchat` DEFAULT CHARACTER SET utf8mb4")

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
		&models.Room{},
		&models.RoomFriend{},
		&models.RoomGroup{},
		&models.GroupMember{},
		&models.Message{},
		&models.UserFriend{},
		&models.UserApply{},
	)
	config.Global.DB = db
}
