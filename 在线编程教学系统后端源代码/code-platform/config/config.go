package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	Mysql        *viper.Viper
	Redis        *viper.Viper
	RedisLRU     *viper.Viper
	Workspace    *viper.Viper
	Minio        *viper.Viper
	Theia        *viper.Viper
	Monaco       *viper.Viper
	IDEServer    *viper.Viper
	MonacoServer *viper.Viper
	Mail         *viper.Viper
)

func init() {
	viper.SetDefault("mysql.host", "127.0.0.1")
	viper.SetDefault("mysql.port", 3306)
	viper.SetDefault("mysql.user", "root")
	viper.SetDefault("mysql.password", "234oix29l2")
	viper.SetDefault("mysql.database", "code_platform")
	viper.SetDefault("mysql.database_test", "code_platform_test")

	viper.SetDefault("redis.host", "127.0.0.1")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "lgbgblwa1243")

	viper.SetDefault("redis_lru.host", "127.0.0.1")
	viper.SetDefault("redis_lru.port", 6380)
	viper.SetDefault("redis_lru.password", "2eobholgbgbl")

	viper.SetDefault("workspace.base_path", map[string]string{
		"windows": "C://code_platform",
		"linux":   "/code_platform/workspace",
	})

	viper.SetDefault("minio.endpoint", "127.0.0.1:9100")
	viper.SetDefault("minio.accessKeyID", "admin")
	viper.SetDefault("minio.secretAccessKey", "12345678")
	viper.SetDefault("minio.bucketName", map[string]string{
		"picture":    "picture",
		"report":     "report",
		"attachment": "attachment",
		"video":      "video",
	})

	minioHost := os.Getenv("MINIO_HOST")
	if minioHost == "" {
		minioHost = "127.0.0.1"
	}
	viper.SetDefault("minio.urlPrefix", minioHost)

	viper.SetDefault("theia.imageName", map[string]string{
		"cpp":     "lgbgbl/theia-cpp-auth",
		"java":    "lgbgbl/theia-java-auth",
		"python3": "lgbgbl/theia-python-auth",
	})

	dockerHost := os.Getenv("THEIA_HOST")
	if dockerHost == "" {
		dockerHost = "127.0.0.1"
	}
	viper.SetDefault("theia.dockerHost", dockerHost)

	viper.SetDefault("monaco.imageName", map[string]string{
		"cpp":     "lgbgbl/monaco-cpp",
		"java":    "lgbgbl/monaco-java",
		"python3": "lgbgbl/monaco-python",
	})

	viper.SetDefault("ide_server.port", 8085)
	viper.SetDefault("monaco_server.port", 8087)

	Mysql = viper.Sub("mysql")
	Redis = viper.Sub("redis")
	RedisLRU = viper.Sub("redis_lru")
	Workspace = viper.Sub("workspace")
	Minio = viper.Sub("minio")
	Theia = viper.Sub("theia")
	Monaco = viper.Sub("monaco")
	IDEServer = viper.Sub("ide_server")
	MonacoServer = viper.Sub("monaco_server")

	Mail = viper.New()
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// 回溯到项目根目录
	for counter := 0; counter < 100; counter++ {
		if filepath.Base(path) == "code-platform" {
			Mail.SetConfigFile(filepath.Join(path, "config", "mail.yaml"))
			Mail.SetConfigType("yaml")
			if err := Mail.ReadInConfig(); err != nil {
				panic(err)
			}
			break
		}
		if path == "/" {
			break
		}
		path = filepath.Dir(path)
	}
}
