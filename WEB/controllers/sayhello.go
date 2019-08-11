package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
)

type SayController struct {
	beego.Controller
}

func (c *SayController) Sayhello() {
	fmt.Println("say hello")
	c.Data["Website"] = "beego.me"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplName = "index.tpl"
	c.Ctx.Output.Header("saywhat", "hello")
}

func (c *SayController) Gethello() {
	c.Ctx.WriteString("<br>hello Gethello!<br/>")
	c.Ctx.Output.Header("saywhat", "Gethello")
	name := c.GetString("name")
	c.Ctx.WriteString(name)
}
