package main

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/z-sdk/goa/tools/goa/mysql/command"
	"os"
)

var (
	BuildTime = "2020.10.10"
	commands  = []cli.Command{
		{
			Name:   "mysql",
			Usage:  "从数据源生成MySQL模型层代码",
			Action: command.GenCodeFromDSN,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dsn",
					Usage: `数据库连接地址，如 "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label"`,
				},
				cli.StringFlag{
					Name:  "table, t",
					Usage: `表名，多表以英文逗号分隔，如 "node,tag,channel"`,
				},
				cli.StringFlag{
					Name:  "dir, d",
					Usage: "目标文件夹",
				},
				cli.BoolFlag{
					Name:  "cache, c",
					Usage: "生成带缓存的数据访问层[可选]",
				},
			},
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Usage = "goa代码生成器"
	app.Version = BuildTime
	app.Commands = commands
	if err := app.Run(os.Args); err != nil {
		fmt.Println("错误：", err)
	}
}
