package main

import (
	"io"

	"yanfei_backend/common"
	"yanfei_backend/controller"

	_ "yanfei_backend/docs"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func migrate(db *gorm.DB) {
	// 后面可以使用AutoMigrate，保持数据库的统一
	// AutoMigration只会根据struct tag建立新表、没有的列以及索引
	// 不会改变已经存在的列的类型或者删除没有用到的列
	db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_bin auto_increment=1")
}

// init 在 main 之前执行
func init() {
	// init config
	common.DefaultConfig()
	common.SetConfig()
	common.WatchConfig()

	// init logger
	common.InitLogger()

	// init Database
	db := common.InitMySQL()
	// 禁止在表名后面加s
	db.SingularTable(true)
	migrate(db)
}

// @title YANFEI API
// @version 0.0.1
func main() {
	// Before init router
	if viper.GetBool("basic.debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
		// Redirect log to file
		gin.DisableConsoleColor()
		logFile := common.GetLogFile()
		defer logFile.Close()
		gin.DefaultWriter = io.MultiWriter(logFile)
	}

	r := gin.Default()
	// middleware
	r.Use(common.ErrorHandling())
	r.Use(common.MaintenanceHandling())

	// swagger router
	if viper.GetBool("basic.debug") {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 路由
	r.GET("/ping", controller.Ping)
	// user相关路由
	r.GET("/user/login", controller.Login)

	r.Run("0.0.0.0:" + viper.GetString("basic.port"))
}
