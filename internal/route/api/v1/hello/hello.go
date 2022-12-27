package hello

import (
	"github.com/NII-DG/gogs/internal/context"
)

type StrContext struct {
	Kotoba string `json:"kotoba" binding:"Required"`
}

type StrKotoba struct {
	Kotoba string `json:"kotoba"`
}

func Hello(c *context.APIContext) {
	c.JSONSuccess("Hello World")
}

func Return(c *context.APIContext) {
	schema := c.QueryEscape("schema")
	//schema := c.Query("schema")
	c.JSONSuccess(schema)
}

func MojiJSON(c *context.APIContext, s StrContext) {
	c.JSONSuccess(s)
}

func Moji(c *context.APIContext, s StrContext) {
	c.JSONSuccess(s.Kotoba)
}

func retKotoba(c *context.APIContext, s StrKotoba) {
	c.JSONSuccess(s)
}
