package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/goredis"
	_ "github.com/go-sql-driver/mysql"
)

type TestModelController struct {
	beego.Controller
}

type UserInfo struct {
	Id       int64
	Username string
	Password string
}

const (
	URL_QUEUE    = "url_queue"
	URL_VIST_SET = "url_vist_set"
)

var (
	redisclient goredis.Client
)

func ConnectRedis(addr string) {
	redisclient.Addr = addr
}

// 放进队列里
func PutinQueue(url string) {
	redisclient.Lpush(URL_QUEUE, []byte(url))
}

// 从队列里取出来
func PopFromQueue() string {
	res, err := redisclient.Rpop(URL_QUEUE)
	if err != nil {
		panic(err)
	}
	return string(res)
}

func AddToSet(url string) {
	redisclient.Sadd(URL_VIST_SET, []byte(url))
}

func ISVist(url string) bool {
	// 判断是否在redis set里存储了
	return true
}

func (c *TestModelController) GoRedisAndQueueTest() {
	ConnectRedis("127.0.0.1:6379")
	// redis 基本操作命令：keys *、 smembers url_queue
	PutinQueue("www.baidu.com")
	c.Ctx.WriteString("redis test success!")
	value := PopFromQueue()
	fmt.Println("test redis value")
	fmt.Println(value)
	c.Ctx.WriteString(value)
}

func (c *TestModelController) HttpLibTest() {
	req := httplib.Get("https://www.baidu.com")
	reqstring, err := req.String()
	if err != nil {
		fmt.Println(reqstring)
	}
	c.Ctx.WriteString("hello httptest")
	// need to test
	c.Ctx.WriteString(reqstring)
}

func (c *TestModelController) TestModel() {
	// beego操作数据库 mysql,数据库表名和对象对应关系
	// AuthUser -> auth_user
	// Auth_User -> auth__user
	// DB_AuthUser -> d_b__auth_user
	// register要写全，没有写?charset=utf8", 30会报错
	orm.RegisterDataBase("default", "mysql", "root:root@tcp(127.0.0.1:3306)/mysql?charset=utf8", 30)
	orm.RegisterModel(new(UserInfo))
	o := orm.NewOrm()
	user := UserInfo{Username: "wangsan", Password: "123456"}
	id, err := o.Insert(&user)
	c.Ctx.WriteString(fmt.Sprintf("test result %d, %v", id, err))
	id, err = o.Update(&UserInfo{Id: 1, Username: "lisi"})
	c.Ctx.WriteString(fmt.Sprintf("id: %d, err: %v", id, err))
	id, err = o.Delete(&UserInfo{Username: "lisi"})
	c.Ctx.WriteString(fmt.Sprintf("id: %d, err: %v", id, err))
	u := UserInfo{Id: 2}
	// 会把id 为2的数据放进u对象里
	o.Read(&u)
	fmt.Println(u.Username)
	var users []UserInfo
	// sql方式查询，推荐
	o.Raw("select * from user_info").QueryRows(&users)
	fmt.Println(users)
	fmt.Println(len(users))
}
