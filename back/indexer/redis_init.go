package indexer

import (
	"context"
	"fmt"
	"time"

	"chainqa_offchain_demo/setting"

	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"github.com/go-redis/redis/v8"
)

// InitRedisClient 初始化 Redis 客户端
func InitRedisClient(ctx context.Context) (*redis.Client, error) {
	redisConf := setting.Conf.Redis

	// 构建 Redis 地址
	addr := fmt.Sprintf("%s:%d", redisConf.Host, redisConf.Port)

	// 创建 Redis 客户端选项
	options := &redis.Options{
		Addr:         addr,
		DB:           redisConf.DB,
		PoolSize:     redisConf.PoolSize,
		DialTimeout:  time.Duration(redisConf.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(redisConf.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(redisConf.WriteTimeout) * time.Second,
	}

	// 如果设置了密码，添加密码选项
	if redisConf.Password != "" {
		options.Password = redisConf.Password
	}

	// 创建客户端
	client := redis.NewClient(options)

	// 测试连接（使用 context.Background()，因为初始化时不需要外部超时控制）
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// NewIndexerService 创建索引服务实例
func NewIndexerService(chainClient *sdk.ChainClient, ctx context.Context) (*IndexerService, error) {
	// 初始化 Redis 客户端
	redisClient, err := InitRedisClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}

	return &IndexerService{
		chainClient: chainClient,
		redisClient: redisClient,
		ctx:         ctx,
	}, nil
}
