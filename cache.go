package redisjsoncache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type Account struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type Cache struct {
	url  string
	rdb  *redis.Client
	ctx  context.Context
	tick *time.Ticker
}

func NewCache(url, redisAddr string, updateInterval time.Duration) *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &Cache{
		url:  url,
		rdb:  rdb,
		ctx:  context.Background(),
		tick: time.NewTicker(updateInterval),
	}
}

func (c *Cache) Start() {
	c.loadAccountsToRedis()

	go func() {
		for range c.tick.C {
			c.loadAccountsToRedis()
		}
	}()
}

func (c *Cache) Stop() {
	c.tick.Stop()
}

func (c *Cache) loadAccountsToRedis() {
	content, err := c.getFileContent(c.url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	accounts, err := c.parseAccounts(content)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	for _, account := range accounts {
		err := c.rdb.Set(c.ctx, "ton_assets:"+account.Address, account.Name, 0).Err()
		if err != nil {
			fmt.Println("Error setting value in Redis:", err)
		}
	}

	fmt.Println("Accounts successfully loaded into Redis")
}

func (c *Cache) getFileContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Cache) parseAccounts(jsonContent string) ([]Account, error) {
	var accounts []Account
	err := json.Unmarshal([]byte(jsonContent), &accounts)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (c *Cache) GetNameByAddress(address string) (string, error) {
	name, err := c.rdb.Get(c.ctx, "ton_assets:"+address).Result()
	if err == redis.Nil {
		return "", errors.New("address not found")
	} else if err != nil {
		return "", err
	}
	return name, nil
}
