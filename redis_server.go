package lantern_cache

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/redcon"
)

const (
	Delimiter = '^'
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
				_ = conn.Close()
			case "set":
				// set key value
				// SET key value EX seconds
				size := len(cmd.Args)
				if size != 3 && size != 5 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}

				var err error
				if size == 3 {
					err = r.cache.Put(cmd.Args[1], cmd.Args[2])
				} else if size == 5 {
					if seconds, err := strconv.ParseInt(string(cmd.Args[4]), 10, 64); err == nil {
						err = r.cache.PutWithExpire(cmd.Args[1], cmd.Args[2], seconds)
					}
				}
				if err != nil {
					conn.WriteError(err.Error())
				} else {
					conn.WriteInt(1)
				}
			case "setex":
				// SETEX key seconds value
				if len(cmd.Args) != 4 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}

				seconds, err := strconv.ParseInt(string(cmd.Args[2]), 10, 64)
				if err != nil {
					conn.WriteError("ERR wrong ttl of argument for" + string(cmd.Args[0]) + "' command")
				}

				err = r.cache.PutWithExpire(cmd.Args[1], cmd.Args[2], seconds)
				if err != nil {
					conn.WriteError(err.Error())
				} else {
					conn.WriteInt(1)
				}
			case "mset":
				// MSET key1 value1 key2 value2 .. keyN valueN
				size := len(cmd.Args)
				if size < 3 || size&1 == 0 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				var err error
				for i := 1; i < size-1; i += 2 {
					if err = r.cache.Put(cmd.Args[i], cmd.Args[i+1]); err != nil {
						break
					}
				}
				if err != nil {
					conn.WriteInt(0)
				} else {
					conn.WriteInt((size - 1) / 2)
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
			case "mget":
				// MGET KEY1 KEY2 .. KEYN
				size := len(cmd.Args)
				if size < 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				conn.WriteArray(size - 1)
				for i := 1; i < size; i++ {
					val, err := r.cache.Get(cmd.Args[i])
					if err != nil {
						conn.WriteNull()
					} else {
						conn.WriteBulk(val)
					}
				}
			case "hset":
				// HSET KEY_NAME FIELD VALUE
				if len(cmd.Args) != 4 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				var key bytes.Buffer
				key.Write(cmd.Args[1])
				key.WriteByte(Delimiter)
				key.Write(cmd.Args[2])
				err := r.cache.Put(key.Bytes(), cmd.Args[3])
				if err != nil {
					conn.WriteError(err.Error())
				} else {
					conn.WriteInt(1)
				}
			case "hmset":
				// HMSET KEY_NAME FIELD1 VALUE1 ...FIELDN VALUEN
				size := len(cmd.Args)
				if size < 3 || size&1 == 1 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}

				var err error
				var key bytes.Buffer
				for i := 2; i < size-1; i += 2 {
					key.Reset()
					key.Write(cmd.Args[1])
					key.WriteByte(Delimiter)
					key.Write(cmd.Args[i])
					if err = r.cache.Put(key.Bytes(), cmd.Args[i+1]); err != nil {
						break
					}
				}
				if err != nil {
					conn.WriteInt(0)
				} else {
					conn.WriteInt((size - 2) / 2)
				}
			case "hget":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				var key bytes.Buffer
				key.Write(cmd.Args[1])
				key.WriteByte(Delimiter)
				key.Write(cmd.Args[2])
				val, err := r.cache.Get(key.Bytes())
				if err != nil {
					conn.WriteNull()
				} else {
					conn.WriteBulk(val)
				}
			case "hmget":
				//HMGET KEY_NAME FIELD1...FIELDN
				size := len(cmd.Args)
				if size < 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}

				var key bytes.Buffer
				conn.WriteArray(size - 2)
				for i := 2; i < size; i++ {
					key.Reset()
					key.Write(cmd.Args[1])
					key.WriteByte(Delimiter)
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
			//fmt.Printf("accept: %s\n", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			fmt.Printf("closed: %s, err: %v\n", conn.RemoteAddr(), err)
		},
	)
	return err
}
