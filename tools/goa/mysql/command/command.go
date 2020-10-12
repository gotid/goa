package command

import (
	"github.com/urfave/cli"
	"github.com/z-sdk/goa/lib/collection"
	"github.com/z-sdk/goa/lib/logx"
	"github.com/z-sdk/goa/lib/store/sqlx"
	"github.com/z-sdk/goa/tools/goa/mysql/gen"
	"github.com/z-sdk/goa/tools/goa/mysql/model"
	"github.com/z-sdk/goa/tools/goa/util"
	"strings"
)

const (
	flagDSN   = "dsn"
	flagTable = "table"
	flagDir   = "dir"
	flagCache = "cache"
)

func GenCodeFromDSN(ctx *cli.Context) error {
	dsn := ctx.String(flagDSN)
	dir := ctx.String(flagDir)
	cache := ctx.Bool(flagCache)
	table := strings.TrimSpace(ctx.String(flagTable))

	if len(dsn) == 0 {
		logx.Error("MySQL连接地址未提供")
		return nil
	}
	if len(table) == 0 {
		logx.Error("表名未提供")
		return nil
	}

	tables := collection.NewSet()
	for _, table := range strings.Split(table, ",") {
		table = strings.TrimSpace(table)
		if len(table) == 0 {
			continue
		}
		tables.AddStr(table)
	}
	logx.Disable()
	conn := sqlx.NewMySQL(dsn)
	m := model.NewModel(conn)
	ddlList, err := m.ShowDDL(tables.KeysStr()...)
	if err != nil {
		logx.Error(err)
		return nil
	}

	//fmt.Println(strings.Join(ddlList, "\n"), dir, cache)
	log := util.NewConsole(true)
	generator := gen.NewModelGenerator(ddlList, dir, gen.WithConsoleOption(log))
	err = generator.Start(cache)
	if err != nil {
		log.Error("", err)
	}

	return nil
}
