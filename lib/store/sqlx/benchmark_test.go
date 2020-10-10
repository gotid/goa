package sqlx

import (
	"goa/lib/logx"
	"testing"
)

func BenchmarkTagQuery(b *testing.B) {
	logx.Disable()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db := NewMySQL("root:asdfasdf@tcp(192.168.0.166:3306)/nest_label")
			result := struct {
				Total int    `conn:"totalx"`
				Name  string `conn:"book"`
			}{}
			err := db.Query(&result, "select book, count(0) totalx from book group by book order by totalx desc")
			if err != nil {
				b.Fatalf("%v", err)
			}
		}
	})
}

func BenchmarkNoTagQuery(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db := NewMySQL("root:asdfasdf@tcp(192.168.0.166:3306)/nest_label")
			result := struct {
				Name  string
				Total int
			}{}
			err := db.Query(&result, "select book, count(0) totalx from book group by book order by totalx desc")
			if err != nil {
				b.Fatalf("%v", err)
			}
		}
	})
}
