package mem

import (
	"github.com/cpu6660/sf-common/conf"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

var RedisTimeOut = 10 * time.Second

var (
	RedisClientsSingle *RedisClients
	redisMutex         sync.Mutex
)

type RedisClients struct {
	clients map[string]*redis.Client
	config  *conf.Config
	sync.Mutex
}

func NewRedisClients(config *conf.Config, single bool) *RedisClients {

	redisMutex.Lock()
	defer redisMutex.Unlock()

	if single && RedisClientsSingle != nil {
		return RedisClientsSingle
	}

	redisClients := &RedisClients{}
	redisClients.clients = make(map[string]*redis.Client)
	redisClients.config = config

	if single {
		RedisClientsSingle = redisClients
	}

	return redisClients

}

func (r RedisClients) GetClient(redisName string) (*redis.Client, error) {

	var (
		err error
	)

	r.Lock()
	defer r.Unlock()

	if _, ok := r.clients[redisName]; ok {
		client := r.clients[redisName]
		err = checkRedisStatus(client)

		if err == nil {
			return r.clients[redisName], nil
		}

	}


	if err != nil {
		delete(r.clients, redisName)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     r.config.GetString(redisName + ":addr"),
		Password: r.config.GetString(redisName + ":password"),
		DB:       r.config.GetInt(redisName + ":db"),
	})

	//check redis status
	err = checkRedisStatus(client)
	if err != nil {
		return nil, err
	}
	r.clients[redisName] = client
	return client, nil
}

func checkRedisStatus(client *redis.Client) error {
	pong, err := client.Ping().Result()
	if (err != nil) || (pong != "PONG") {
		return err
	}
	return nil
}
