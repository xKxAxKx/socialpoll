package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// *http.Requestオブジェクトからのデコード
func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	// https://xn--go-hh0g6u.com/pkg/encoding/json/#Decoder.Decode
	return json.NewDecoder(r.Body).Decode(v)
}

// ResponseWriterオブジェクトへのエンコード
func encodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

/*
  レスポンスの生成を行う関数。
  encodeBodyを使ってエンコードしたデータが、
  ステータスコードと合わせてResponseWriteに出力される
*/
func respond(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		encodeBody(w, r, data)
	}
}

/*
  エラーレスポンスの生成を行う関数。
  問題の発生であ ることを明示するために、errorオブジェクトのmessageとして出力
*/
func respondErr(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	args ...interface{},
) {
	respond(w, r, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": fmt.Sprint(args...),
		},
	})
}

/*
  HTTP でのエラーに特化したヘルパー
  Go の標準ライブラリに含まれる http.StatusText 関数を使い、適切なメッセージを生成
*/
func respondHTTPErr(w http.ResponseWriter, r *http.Request,
	status int,
) {
	respondErr(w, r, status, http.StatusText(status))
}
