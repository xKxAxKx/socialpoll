package main

import (
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
  reflect パッケージを使うと、タグとして記述されたキーと値の組にアクセスできる
  MongoDB とのやり取りには BSON が使われ、クライアントに対しては JSON が使われる
*/
type poll struct {
	ID      bson.ObjectId  `bson:"_id" json:"id"`
	Title   string         `json:"title"`
	Options []string       `json:"options"`
	Results map[string]int `json:"results,omitempty"`
}

/*
  HTTPメソッドの値に対するswitch文の中で、
  この値がGETやPOSTそしてDELETEの場合の処理をそれぞれ記述
*/
func handlePolls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlePollsGet(w, r)
		return
	case "POST":
		handlePollsPost(w, r)
		return
	case "DELETE":
		handlePollsDelete(w, r)
		return
	case "OPTIONS":
		// CORS に対応したブラウザは、DELETE のリク エストに先立って
		// pre-flight リクエストと呼ばれる OPTIONS 形式のリクエストを行い
		// 実際のリクエストを行うための許可を求める
		// それを適切に対応するに必要
		w.Header().Add("Access-Control-Allow-Methods", "DELETE")
		respond(w, r, http.StatusOK, nil)
		return
	}
	// 未対応のHTTPメソッド
	respondHTTPErr(w, r, http.StatusNotFound)
}

func handlePollsGet(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("polls")
	var q *mgo.Query

	// NewPathを用いてリクエストURLのパース
	p := NewPath(r.URL.Path)
	if p.HasID() {
		// 特定の調査項目の詳細を取得する
		q = c.FindId(bson.ObjectIdHex(p.ID))
	} else {
		// すべての調査項目のリストを取得
		q = c.Find(nil)
	}
	var result []*poll
	if err := q.All(&result); err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, &result)
}

func handlePollsPost(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("polls")
	var p poll
	// decodeBody()で、pにリクエストボディのデータが埋め込まれいるっぽいな...
	if err := decodeBody(r, &p); err != nil {
		respondErr(w, r, http.StatusBadRequest, "リクエストから調査項目を読み込めません", err)
		return
	}
	p.ID = bson.NewObjectId()
	if err := c.Insert(p); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "調査項目の格納に失敗しました", err)
		return
	}
	w.Header().Set("Location", "polls/"+p.ID.Hex())
	respond(w, r, http.StatusCreated, nil)
}

func handlePollsDelete(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("polls")
	p := NewPath(r.URL.Path)
	if !p.HasID() {
		respondErr(w, r, http.StatusMethodNotAllowed, "すべての調査項目を削除することはできません")
		return
	}
	if err := c.RemoveId(bson.ObjectIdHex(p.ID)); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "調査項目の削除に失敗しました", err)
		return
	}
	respond(w, r, http.StatusOK, nil) // 成功
}
