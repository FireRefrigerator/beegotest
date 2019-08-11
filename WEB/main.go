package main

import (
	_ "WEB/routers"
	"fmt"

	"github.com/astaxie/beego"
)

func main() {
	fmt.Println("beego run")
	beego.Run()
}
