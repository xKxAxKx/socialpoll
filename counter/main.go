package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	nsq "github.com/bitly/go-nsq"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var fatalErr error

// 問題が発生したら呼び出される関数
func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

const updateDuration = 1 * time.Second

func main() {
	// main 関数の終了後に defer しておいたコードが呼び出され、その中で終了コード 1 を返してプログラム全体が終了
	// deferは後入れ先出しなので、最初に書いてたやつが最後に処理される
	// どんなエラーが発生してもDB接続を必ず閉じたいためこういう書き方にしている
	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	log.Println("データベースに接続します...")
	db, err := mgo.Dial("localhost")
	if err != nil {
		fatal(err)
		return
	}

	// 上記のdeferよりもこちらが先に実行される
	defer func() {
		log.Println("データベース接続を閉じます...")
		db.Close()
	}()
	pollData := db.DB("ballots").C("polls")

	// マップとロック(sync.Mutex)はGoでよく使われる組み合わせ
	// 複数のゴルーチンが1つのマップのアクセスにする際に、
	// 同時に読み書きを行なってマップを破壊するのを防ぐ
	var countsLock sync.Mutex
	var counts map[string]int

	log.Println("NSQに接続します...")
	// NSQ の votes トピックを監視するオブジェクトがセットアップされる
	// twittervotesがパブリッシュしたメッセージを読み出せるようになる
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		fatal(err)
		return
	}

	// この関数は、votes上でメッセージが受信されるたびに実行される
	// pub.Publish("votes", []byte(vote)) <- twittervotes/main.go内のここ？
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		// 他のプロセスがマップcountsを読み書きできないようにcountsLockがロックされる
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil {
			counts = make(map[string]int)
		}
		vote := string(m.Body)
		// マップcountsの更新
		counts[vote]++
		return nil
	}))

	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil {
		fatal(err)
		return
	}

	log.Println("NSQ上での投票を待機します...")
	var updater *time.Timer
	// 引数として指定された関数を一定時間後に自身のgoroutineの中で実行する
	updater = time.AfterFunc(updateDuration, func() {
		countsLock.Lock()
		defer countsLock.Unlock()
		if len(counts) == 0 {
			log.Println("新しい投票はありません。データベースの更新をスキップします")
		} else {
			log.Println("データベースを更新します...")
			log.Println(counts)
			ok := true
			for option, count := range counts {
				sel := bson.M{"options": bson.M{"$in": []string{option}}}
				up := bson.M{"$inc": bson.M{"results." + option: count}}
				if _, err := pollData.UpdateAll(sel, up); err != nil {
					log.Println("更新に失敗しました:", err)
					ok = false
					continue
				}
				counts[option] = 0
			}
			if ok {
				log.Println("データベースの更新が完了しました")
				counts = nil // 得票数をリセットします
			}
		}
		// Reset を呼び出し、同じ手順を再び行う
		// つまり、更新のためのコード が定期的に繰り返し実行される
		updater.Reset(updateDuration)
	})

	// Ctrl + C が押された時に発生する終了のイベントを捕捉するためのチャネル
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		// select文を使って termChan と nsq.ConsumerのStopChan へのメッセージを待機
		select {
		// Ctrl + C が押された際には、まず termChan にシグナルが送られる
		// updaterのタイマーが停止され、Consumerに対して投票への監視を停止するよう指示される
		// そしてループが再開され、Consumerが完了して自身のStopChanにシグナルを送るまで実行はブロックします。
		case <-termChan:
			updater.Stop()
			q.Stop()
		case <-q.StopChan:
			// 完了しました
			return
		}
	}
}
