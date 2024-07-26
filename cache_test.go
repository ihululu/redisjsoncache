package redisjsoncache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	url := "https://raw.githubusercontent.com/tonkeeper/ton-assets/main/accounts.json"
	redisAddr := "localhost:6379"
	updateInterval := 24 * time.Hour

	cache := NewCache(url, redisAddr, updateInterval)
	cache.Start()
	defer cache.Stop()

	time.Sleep(1 * time.Second) // 等待初次加载完成

	address := "0:c21fabd3d0cbd5a746c160d8e838da6481be7534c3cd3895c54161b9e17a0802"
	name, err := cache.GetNameByAddress(address)
	if err != nil {
		t.Fatal("Error:", err)
	}

	if name == "" {
		t.Fatal("Expected non-empty name")
	}

	t.Log("Name:", name)
}
