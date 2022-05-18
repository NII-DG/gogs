package db_test

import (
	"log"
	"testing"

	"github.com/NII-DG/gogs/internal/db"
	"xorm.io/xorm"
)

func GetXorm() *xorm.Engine {
	url := "user=gogs host=localhost port=5432 dbname=gogs sslmode=disable"
	engine, err := xorm.NewEngine("postgres", url)
	if err != nil {
		log.Fatalf("データベースの接続に失敗しました。: %v", err)
	}
	return engine
}

func TestBenchPublicUploadRepoFiles(b testing.B) {
	e := GetXorm()
	repository := db.Repository{}
	result, err := e.Where("name = ?", "dddd").Get(&repository)
	if err != nil {
		log.Fatal(err)
	}
	if !result {
		log.Fatal("Not Found")
	}

}
