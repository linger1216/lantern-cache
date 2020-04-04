package lantern_cache

import (
	"bytes"
	"fmt"
	"github.com/tidwall/redcon"
	"strconv"
	"strings"
)

type RedisServer struct {
	addr  string
	cache *LanternCache
}

func NewRedisServer(addr string, cache *LanternCache) *RedisServer {
	return &RedisServer{addr: addr, cache: cache}
}

func (r *RedisServer) ListenAndServe() error {
	err := redcon.ListenAndServe(r.addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			switch strings.ToLower(string(cmd.Args[0])) {
			default:
				conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
			case "select":
				conn.WriteString("OK")
			case "ping":
				conn.WriteString("pong")
			case "quit", "exit":
				conn.WriteString("OK")
				conn.Close()
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				err := r.cache.Put(cmd.Args[1], cmd.Args[2])
				if err != nil {
					conn.WriteError(err.Error())
				} else {
					conn.WriteInt(1)
				}
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				val, err := r.cache.Get(cmd.Args[1])
				if err != nil {
					conn.WriteNull()
				} else {
					conn.WriteBulk(val)
				}
			case "hset":
				if len(cmd.Args) != 4 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				var key bytes.Buffer
				key.Write(cmd.Args[1])
				key.WriteByte('#')
				key.Write(cmd.Args[2])
				err := r.cache.Put(key.Bytes(), cmd.Args[3])
				if err != nil {
					conn.WriteError(err.Error())
				} else {
					conn.WriteInt(1)
				}
			case "hmset":
				bucket := cmd.Args[1]
				for i := 2; i < len(cmd.Args)-1; i += 2 {
					var key bytes.Buffer
					key.Write(bucket)
					key.WriteByte('#')
					key.Write(cmd.Args[i])
					err := r.cache.Put(key.Bytes(), cmd.Args[i+1])
					if err != nil {
						conn.WriteNull()
					} else {
						conn.WriteInt((len(cmd.Args) - 2) / 2)
					}
				}
			case "hget":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				var key bytes.Buffer
				key.Write(cmd.Args[1])
				key.WriteByte('#')
				key.Write(cmd.Args[2])
				val, err := r.cache.Get(key.Bytes())
				if err != nil {
					conn.WriteNull()
				} else {
					conn.WriteBulk(val)
				}
			case "hmget":
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				conn.WriteArray(len(cmd.Args) - 2)
				for i := 2; i < len(cmd.Args); i++ {
					var key bytes.Buffer
					key.Write(cmd.Args[1])
					key.WriteByte('#')
					key.Write(cmd.Args[i])
					val, err := r.cache.Get(key.Bytes())
					if err != nil {
						conn.WriteNull()
					} else {
						conn.WriteBulk(val)
					}
				}
			case "del":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				r.cache.Del(cmd.Args[1])
				conn.WriteInt(1)
			case "dbsize":
				conn.WriteUint64(r.cache.Size())
			case "scan":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				count, _ := strconv.ParseInt(string(cmd.Args[1]), 10, 64)
				if count == 0 {
					count = 100
				}
				ret, err := r.cache.Scan(int(count))
				if err != nil {
					conn.WriteError(err.Error())
				} else {
					conn.WriteArray(len(ret))
					for i := range ret {
						conn.WriteBulk(ret[i])
					}
				}
			}
		},
		func(conn redcon.Conn) bool {
			fmt.Printf("accept: %s\n", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			fmt.Printf("closed: %s, err: %v\n", conn.RemoteAddr(), err)
		},
	)
	return err
}
