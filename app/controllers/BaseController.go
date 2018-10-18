package controllers

import (
	"github.com/kataras/iris"
)

type Base struct{}

func (Base) Index(ctx iris.Context) {
	ctx.JSON(iris.Map{"message": "Hello Iris!"})
}
