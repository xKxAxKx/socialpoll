package main

import (
	"flag"
	"fmt"
	"os"
)

var fatalErr error

// 問題が発生したら呼び出される関数
func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

func main() {
	// main 関数の終了後に defer しておいたコードが呼び出され、その中で終了コード 1 を返してプログラム全体が終了
	// deferは後入れ先出しなので、最初に書いてたやつが最後に処理される
	// どんなエラーが発生してもDB接続を必ず閉じたいためこういう書き方にしている
	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()
}
