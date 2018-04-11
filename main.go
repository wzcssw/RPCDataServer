// go run -tags etcd main.go

package main

import (
	"context"
	"flag"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

var (
	addr     = flag.String("addr", "localhost:8972", "server address")
	etcdAddr = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath = flag.String("base", "/rpcx_users", "prefix path")
	DB       *gorm.DB
)

type Users struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	Name   string
	Gender int
	Mark   string
}

func (self *Users) GetAllUsers(ctx context.Context, args *Users, reply *[]Users) error {
	DB.Find(reply)
	return nil
}

func (self *Users) GetUser(ctx context.Context, args *Users, reply *Users) error {
	DB.Where("id = ?", args.ID).First(reply)
	return nil
}

func (self *Users) AddUser(ctx context.Context, args *Users, reply *Users) error {
	DB.Create(&args)
	return nil
}

func (self *Users) UpdateUser(ctx context.Context, args *Users, reply *Users) error {
	DB.Save(&args)
	return nil
}

func init() {
	DB, _ = gorm.Open("mysql", "root:wzc19931030@tcp(127.0.0.1:3306)/funny?charset=utf8&parseTime=true")
}

func main() {
	s := server.NewServer()
	addRegistryPlugin(s)
	s.RegisterName("Users", new(Users), "")
	s.Serve("tcp", *addr)
	defer s.Close()
}

func addRegistryPlugin(s *server.Server) {
	r := &serverplugin.EtcdRegisterPlugin{
		ServiceAddress: "tcp@" + *addr,
		EtcdServers:    []string{*etcdAddr},
		BasePath:       *basePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}
