package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/lock"
	"github.com/yi-nology/git-manage-service/pkg/storage"
)

func main() {
	fmt.Println("=== Git Manage Service - Redis & MinIO 测试 ===")
	fmt.Println()

	// 加载测试配置
	cfg, err := configs.LoadConfig([]string{".", "./conf"}, "config_test", "yaml")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 测试 Redis 锁
	fmt.Println("--- 测试 Redis 分布式锁 ---")
	testRedisLock(cfg.Lock)

	fmt.Println()

	// 测试 MinIO 存储
	fmt.Println("--- 测试 MinIO 对象存储 ---")
	testMinIOStorage(cfg.Storage)

	fmt.Println()
	fmt.Println("=== 测试完成 ===")
}

func testRedisLock(cfg configs.LockConfig) {
	fmt.Printf("锁类型: %s\n", cfg.Type)
	fmt.Printf("Redis 地址: %s\n", cfg.RedisAddr)

	lockSvc, err := lock.NewDistLock(cfg)
	if err != nil {
		fmt.Printf("创建锁服务失败: %v\n", err)
		return
	}
	defer lockSvc.Close()

	ctx := context.Background()
	testKey := "test:lock:key"

	// 测试获取锁
	fmt.Println("\n1. 测试获取锁...")
	success, err := lockSvc.Up(ctx, testKey, 10*time.Second)
	if err != nil {
		fmt.Printf("   获取锁失败: %v\n", err)
		return
	}
	fmt.Printf("   获取锁结果: %v\n", success)

	// 测试重复获取（应该失败）
	fmt.Println("\n2. 测试重复获取锁（应该失败）...")
	success2, err := lockSvc.Up(ctx, testKey, 10*time.Second)
	if err != nil {
		fmt.Printf("   获取锁失败: %v\n", err)
	} else {
		fmt.Printf("   重复获取锁结果: %v (预期: false)\n", success2)
	}

	// 测试释放锁
	fmt.Println("\n3. 测试释放锁...")
	err = lockSvc.Down(ctx, testKey)
	if err != nil {
		fmt.Printf("   释放锁失败: %v\n", err)
	} else {
		fmt.Println("   释放锁成功")
	}

	// 测试释放后再次获取
	fmt.Println("\n4. 测试释放后再次获取...")
	success3, err := lockSvc.Up(ctx, testKey, 10*time.Second)
	if err != nil {
		fmt.Printf("   获取锁失败: %v\n", err)
	} else {
		fmt.Printf("   再次获取锁结果: %v (预期: true)\n", success3)
	}

	// 清理
	lockSvc.Down(ctx, testKey)
	fmt.Println("\nRedis 锁测试通过!")
}

func testMinIOStorage(cfg configs.StorageConfig) {
	fmt.Printf("存储类型: %s\n", cfg.Type)
	fmt.Printf("MinIO 端点: %s\n", cfg.Endpoint)

	storageSvc, err := storage.NewStorage(cfg)
	if err != nil {
		fmt.Printf("创建存储服务失败: %v\n", err)
		return
	}
	defer storageSvc.Close()

	ctx := context.Background()
	testBucket := "test-bucket"
	testKey := "test/hello.txt"
	testContent := "Hello, MinIO! 这是一个测试文件。时间: " + time.Now().Format(time.RFC3339)

	// 测试创建桶
	fmt.Println("\n1. 测试创建桶...")
	err = storageSvc.MakeBucket(ctx, testBucket)
	if err != nil {
		fmt.Printf("   创建桶失败: %v\n", err)
		return
	}
	fmt.Printf("   创建桶 '%s' 成功\n", testBucket)

	// 测试检查桶是否存在
	fmt.Println("\n2. 测试检查桶是否存在...")
	exists, err := storageSvc.BucketExists(ctx, testBucket)
	if err != nil {
		fmt.Printf("   检查桶失败: %v\n", err)
	} else {
		fmt.Printf("   桶存在: %v\n", exists)
	}

	// 测试上传对象
	fmt.Println("\n3. 测试上传对象...")
	content := []byte(testContent)
	err = storageSvc.PutObject(ctx, testBucket, testKey, bytes.NewReader(content), int64(len(content)))
	if err != nil {
		fmt.Printf("   上传对象失败: %v\n", err)
		return
	}
	fmt.Printf("   上传对象 '%s' 成功 (%d 字节)\n", testKey, len(content))

	// 测试获取对象元数据
	fmt.Println("\n4. 测试获取对象元数据...")
	info, err := storageSvc.StatObject(ctx, testBucket, testKey)
	if err != nil {
		fmt.Printf("   获取元数据失败: %v\n", err)
	} else {
		fmt.Printf("   对象大小: %d 字节\n", info.Size)
		fmt.Printf("   最后修改: %s\n", info.LastModified.Format(time.RFC3339))
	}

	// 测试下载对象
	fmt.Println("\n5. 测试下载对象...")
	reader, err := storageSvc.GetObject(ctx, testBucket, testKey)
	if err != nil {
		fmt.Printf("   下载对象失败: %v\n", err)
		return
	}
	downloadedContent, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		fmt.Printf("   读取内容失败: %v\n", err)
	} else {
		fmt.Printf("   下载内容: %s\n", string(downloadedContent))
		if string(downloadedContent) == testContent {
			fmt.Println("   内容验证通过!")
		}
	}

	// 测试列出对象
	fmt.Println("\n6. 测试列出对象...")
	objects, err := storageSvc.ListObjects(ctx, testBucket, "")
	if err != nil {
		fmt.Printf("   列出对象失败: %v\n", err)
	} else {
		fmt.Printf("   找到 %d 个对象:\n", len(objects))
		for _, obj := range objects {
			fmt.Printf("   - %s (%d 字节)\n", obj.Key, obj.Size)
		}
	}

	// 测试删除对象
	fmt.Println("\n7. 测试删除对象...")
	err = storageSvc.DeleteObject(ctx, testBucket, testKey)
	if err != nil {
		fmt.Printf("   删除对象失败: %v\n", err)
	} else {
		fmt.Printf("   删除对象 '%s' 成功\n", testKey)
	}

	fmt.Println("\nMinIO 存储测试通过!")
}
