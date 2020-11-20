package redis

// 函数说明：https://www.jianshu.com/p/5fab80fee876
import (
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisPool RedisPool
type RedisPool struct {
	pool *redis.Pool
	conn redis.Conn
}

func (r *RedisPool) newRedisPool() *redis.Pool {
	//代码来源，仅此函数：https://blog.csdn.net/weixin_44540711/article/details/105269317
	return &redis.Pool{
		MaxIdle:     3,                 //最大空闲数
		MaxActive:   100,               //最大活跃数
		IdleTimeout: 240 * time.Second, //最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
		Dial: func() (redis.Conn, error) {
			//此处对应redis ip及端口号
			conn, err := redis.Dial("tcp", "192.168.10.104:6379")
			r.failOnErr(err, "连接Redis错误!")
			//redis密码
			_, err = conn.Do("AUTH", "123456")
			if err != nil {
				conn.Close()
				r.failOnErr(err, "Redis密码错误!")
			}
			return conn, err
		},
	}
}

// NewRedisPool NewRedisPool
func NewRedisPool() *RedisPool {
	Redis := RedisPool{}
	Redis.pool = Redis.newRedisPool()
	Redis.conn = Redis.pool.Get()
	return &Redis
}

// Set 增加值
func (r *RedisPool) Set(name string, value interface{}) {
	_, err := r.conn.Do("set", name, value)
	r.failOnErr(err, "Set Error!")
}

// SetList 增加列表
func (r *RedisPool) SetList(name string, value interface{}) {
	_, err := r.conn.Do("rpush", name, value)
	r.failOnErr(err, "SetList Error!")
}

// Get 取值
func (r *RedisPool) Get(name string) string {
	str, err := redis.String(r.conn.Do("get", name))
	r.failOnErr(err, "Get Error!")
	return str
}

// Close 关闭
func (r *RedisPool) Close() {
	r.conn.Close()
	r.pool.Close()
}

//错误处理函数
func (r *RedisPool) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

//测试
func main() {
	Redis := NewRedisPool()
	defer Redis.Close()
	Redis.Set("name:json:a2", "123456")
	str := Redis.Get("name:json:a2")

	fmt.Println(str)

}
