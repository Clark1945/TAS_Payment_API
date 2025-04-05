package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	result := make(chan string)

	// 建立 context，1 秒 timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel() // 確保最後釋放資源

	go doWork(ctx, result)

	select {
	case res := <-result:
		fmt.Println(res)
	case <-ctx.Done(): // 有兩個地方監聽ctx，感覺寫得不好
		fmt.Println("❌ 超時沒收到結果：", ctx.Err())
	}
}

// func typicalChan() {
// 	// 宣告channel make(chan 型態 <容量>)
// 	val := make(chan int)
// 	// 執行第一個執行緒
// 	go func() {
// 		val <- 1 //注入資料1
// 	}()
// 	// 執行第二個執行緒
// 	go func() {
// 		val <- 2 //注入資料2
// 		time.Sleep(time.Millisecond * 100)
// 	}()
// 	ans := []int{}
// 	for {
// 		ans = append(ans, <-val) //取出資料
// 		fmt.Println(ans)
// 		if len(ans) == 2 {
// 			break
// 		}
// 	}
// }

// func typicalWaitGroup() {
// 	var wg sync.WaitGroup
// 	// 執行執行緒
// 	go func() {
// 		defer wg.Done() //defer表示最後執行，因此該行為最後執行wg.Done()將計數器-1
// 		defer log.Println("goroutine drop out")
// 		log.Println("start a go routine")
// 		time.Sleep(time.Second) //休息一秒鐘
// 	}()
// 	wg.Add(1)                         //計數器+1
// 	time.Sleep(time.Millisecond * 30) //休息30 ms
// 	log.Println("wait a goroutine")
// 	wg.Wait() //等待計數器歸0
// 	fmt.Println("Finish")
// }

// func raceWithMutexLock() {
// 	var lock sync.Mutex   // 宣告Lock 用以資源佔有與解鎖
// 	var wg sync.WaitGroup // 宣告WaitGroup 用以等待執行序
// 	val := 0
// 	// 執行 執行緒: 將變數val+1
// 	go func() {
// 		defer wg.Done() //wg 計數器-1
// 		//使用for迴圈將val+1
// 		for i := 0; i < 10; i++ {
// 			lock.Lock() //佔有資源
// 			val++       // 他被lock住囉
// 			fmt.Printf("First gorutine val++ and val = %d\n", val)
// 			lock.Unlock() //釋放資源
// 			time.Sleep(3000)
// 		}
// 	}()
// 	// 執行 執行緒: 將變數val+1
// 	go func() {
// 		defer wg.Done() //wg 計數器-1
// 		//使用for迴圈將val+1
// 		for i := 0; i < 10; i++ {
// 			lock.Lock() //佔有資源
// 			val++
// 			fmt.Printf("Sec gorutine val++ and val = %d\n", val)
// 			lock.Unlock() // 釋放資源
// 			time.Sleep(1000)
// 		}
// 	}()
// 	wg.Add(2) //記數器+2
// 	wg.Wait() //等待計數器歸零
// }

// func doTheFirstThread() {
// 	firstRoutine := make(chan string) //宣告給第1個執行序的channel
// 	secRoutine := make(chan string)   //宣告給第2個執行序的channel
// 	rand.Seed(time.Now().UnixNano())

// 	go func() {
// 		r := rand.Intn(100)
// 		time.Sleep(time.Microsecond * time.Duration(r)) //隨機等待 0~100 ms
// 		firstRoutine <- "first goroutine"
// 	}()
// 	go func() {
// 		r := rand.Intn(100)
// 		time.Sleep(time.Microsecond * time.Duration(r)) //隨機等待 0~100 ms
// 		secRoutine <- "Sec goroutine"
// 	}()
// 	select { // select多路複用 select 會阻塞等待直到有任一個 case 的 channel 收到資料。
// 	case f := <-firstRoutine: //第1個執行序先執行後所要做的動作
// 		fmt.Println(f)
// 		return
// 	case s := <-secRoutine: // 第2個執行序先執行後所要做的動作
// 		fmt.Println(s)
// 		return
// 	}
// }

func doWork(ctx context.Context, result chan<- string) {
	// 模擬一個耗時不確定的工作
	r := rand.Intn(3000) // 0~2999 毫秒
	select {
	case <-time.After(time.Duration(r) * time.Millisecond): // 延遲
		result <- fmt.Sprintf("✅ 工作完成於 %d 毫秒", r)
	case <-ctx.Done():
		fmt.Println("⛔ 工作被取消，原因：", ctx.Err()) // context timeout 觸發
	}
}
