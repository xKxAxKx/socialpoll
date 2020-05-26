package main

import (
	"net/http"

	mgo "gopkg.in/mgo.v2"
)

func main() {}

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
