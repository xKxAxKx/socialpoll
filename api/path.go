/*
  パスの解析を行うパッケージ
*/
package main

import (
	"strings"
)

const PathSeparator = "/"

type Path struct {
	Path string
	ID   string
}

/*
  受け取ったパスの文字列を解析
  Path型のインスタンスを生成して返す
*/
func NewPath(p string) *Path {
	var id string
	// 先頭と末尾のスラッシュを削除
	p = strings.Trim(p, PathSeparator)

	// スラッシュを区切ってスライスに分割
	s := strings.Split(p, PathSeparator)
	if len(s) > 1 {
		// 最後の項目をidとして取り出す
		// "/fuga/hoge/"みたいに末尾idじゃないエンドポイントの場合おかしくなるけどいいのか？
		id = s[len(s)-1]

		// 末尾以外を新しい文字列として接続
		p = strings.Join(s[:len(s)-1], PathSeparator)
	}
	return &Path{Path: p, ID: id}
}

/*
  NewPathによって生成されたPathがidを持っているか否かを返す
*/
func (p *Path) HasID() bool {
	return len(p.ID) > 0
}
