package testx

import (
	"fmt"
	"time"

	"code-platform/config"
	"code-platform/storage"

	"github.com/jmoiron/sqlx"
)

func NewStorage() *storage.Storage {
	s := storage.NewStorage()
	s.RDB = mustInitMysqlTestingClient()

	// 切换到　Redis 4号测试DB
	s.Redis = mustInitRedisTestingClient(redisTestDBNumber, config.Redis)
	s.RedisLRU = mustInitRedisTestingClient(redisTestDBNumber, config.RedisLRU)

	return s
}

func mustInitMysqlTestingClient() storage.RDBClient {
	user := config.Mysql.GetString("user")
	password := config.Mysql.GetString("password")
	host := config.Mysql.GetString("host")
	port := config.Mysql.GetInt("port")
	dataBase := config.Mysql.GetString("database_test")

	dataSource := fmt.Sprintf(`%s:%s@tcp(%s:%d)/%s?parseTime=true`, user, password, host, port, dataBase)
	dataSource += "&loc=Asia%2FShanghai"
	db, err := sqlx.Connect("mysql", dataSource)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(2000)
	// wait_time ( 28800 s ) -10s
	db.SetConnMaxLifetime(28790 * time.Second)

	db.SetMaxIdleConns(1000)
	// 半小时空闲连接不被访问则被关闭
	db.SetConnMaxIdleTime(time.Minute * 30)
	return db
}
