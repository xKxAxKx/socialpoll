package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/tylerstillwater/graceful"
	mgo "gopkg.in/mgo.v2"
)

func main() {
	/*
	   コマンドライン引数の定義
	   flag.Parse()で変数に取得できる
	   参考:https://qiita.com/Yaruki00/items/7edc04720a24e71abfa2#%E3%83%95%E3%83%A9%E3%82%B0%E3%81%AE%E5%8F%96%E5%BE%97
	*/
	var (
		addr  = flag.String("addr", ":8080", "エンドポイントのアドレス")
		mongo = flag.String("mongo", "localhost", "MongoDBのアドレス")
	)
	flag.Parse()

	log.Println("MongoDBに接続します", *mongo)
	db, err := mgo.Dial(*mongo)
	if err != nil {
		log.Fatalln("MongoDBへの接続に失敗しました:", err)
	}
	defer db.Close()

	/*
	   ServeMux構造体を生成する
	   https://golang.org/pkg/net/http/#ServeMux
	   ServeMux は登録されたパターンのリストと各受信リクエストのURLをマッチさせ
	   URL に最も近いパターンのハンドラを呼び出す。
	*/
	mux := http.NewServeMux()
	// パスが/polls/で始まるリクエストを処理する
	mux.HandleFunc("/polls/",
		withCORS(
			withVars(
				withData(db, withAPIKey(handlePolls)),
			),
		),
	)
	log.Println("Webサーバーを開始します:", *addr)

	// https://github.com/tylerstillwater/graceful
	// http.Handler(ServeMuxも含)の実行時間をtime.Durationとして指定できる
	graceful.Run(*addr, 1*time.Second, mux)
	log.Println("停止します...")
}

// http.HandlerFuncをラップした関数
// 引数も戻り値も http.HandlerFunc型
// -> コンテキストの中にラップする
func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidAPIKey(r.URL.Query().Get("key")) {
			respondErr(w, r, http.StatusUnauthorized, "不正なAPIキーです")
			return
		}
		fn(w, r)
	}
}

func isValidAPIKey(key string) bool {
	// APIキーはabc123とハードコードしておく
	// これ以外が渡されたらfalseを返す
	return key == "abc123"
}

// MongoDB のセッションを表す値(mgo パッケージに含まれる)と、
// 同じパターンに基づいた別のハンドラとを引数として受け取る
func withData(d *mgo.Session, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// まずデータベースセッションの値をコピー
		thisDb := d.Copy()
		defer thisDb.Close()
		// ヘルパー関数 SetVarを使い、ballotsデータベースへの参照をdb変数にセット
		SetVar(r, "db", thisDb.DB("ballots"))
		// 最後に、次のHandlerFuncを呼び出す
		f(w, r)
	}
}

func withVars(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		OpenVars(r)
		defer CloseVars(r)
		fn(w, r)
	}
}

func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		fn(w, r)
	}
}
