package main

import (
	"net/http"
	"sync"
)

var (
	varsLock sync.RWMutex
	// キーはリクエストを表す *http.Request 型で、値は別のマップです
	// 値のマップにはリクエストのインスタンスに関連づけたデータが格納される
	vars map[*http.Request]map[string]interface{}
)

func OpenVars(r *http.Request) {
	// マップを安全に操作できるようにvarsLockのロックを獲得
	varsLock.Lock()
	// varがnilの場合にマップを生成する
	if vars == nil {
		// 指定された http.Request へのポインタをキーとして空のマップをvarsに追加
		vars = map[*http.Request]map[string]interface{}{}
	}
	vars[r] = map[string]interface{}{}
	varsLock.Unlock()
}

// 1 つのリクエストの処理が終わった際に、そこで使われていたメモリを解放してメモリリークを防ぐ
func CloseVars(r *http.Request) {
	varsLock.Lock()
	// 指定されたリクエストに対応するエントリがマップ vars から安全に削除される
	delete(vars, r)
	varsLock.Unlock()
}

// 指定されたリクエストに関連づけられたデータを容易に取得できるようにする
func GetVar(r *http.Request, key string) interface{} {
	// RLock/RUnlock(sync.RWMutex)を使うことにより、書き込みが発生していないかぎり複数の読み出しを同時に行える
	// RLock では、他のコードによる RLock をブロック することはない
	varsLock.RLock()
	value := vars[r][key]
	varsLock.RUnlock()
	return value
}

func SetVar(r *http.Request, key string, value interface{}) {
	varsLock.Lock()
	vars[r][key] = value
	varsLock.Unlock()
}
