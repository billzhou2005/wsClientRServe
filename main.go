// websocket client.go
package main

import (
	"math/rand"
	"wsClientRServe/csmapi"

	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const VOL_TABLE_MAX int = 3
const TABLE_PLAYERS_MAX int = 9

var done chan interface{}
var interrupt chan os.Signal
var sendChan chan [9]rcvMessage

type rcvMessage struct {
	TableID     int    `json:"tableID"`
	UserID      string `json:"userID"`
	Status      string `json:"status"` //system auto flag
	ConnType    string `json:"connType"`
	IsActivated bool   `json:"isActivated"`
	Round       int    `json:"round"`
	SeatID      int    `json:"seatID"`
	Betvol      int    `json:"betvol"`
	Greeting    string `json:"greeting"`
}

/*
func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}
*/
func main() {

	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully
	sendChan = make(chan [9]rcvMessage)

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	players := csmapi.GetPlayersCards(50000012, 6)
	fmt.Println(players)

	socketUrl := "ws://140.143.149.188:9080" + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveJsonHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		// case <-time.After(time.Duration(1) * time.Second * 60):
		case sendMsg := <-sendChan:

			// Send an next player packet if needed
			err := conn.WriteJSON(sendMsg)
			if err != nil {
				log.Println("Error during writing the json data to websocket:", err)
				return
			}
			// err := conn.WriteMessage(websocket.TextMessage, []byte("Test message from Golang ws client every 60 secs"))
			// if err != nil {
			//	log.Println("Error during writing to websocket:", err)
			//	return
			// }

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}
}

// connType: RESERVED,JOINED,ASSINGED,WAITING,ACTIVATE,BNEXT,TIMEOUT,CLOSE
// Create Next player info for sending
func createNextPlayerMsg(usersMsg [9]rcvMessage, seatID int) rcvMessage {
	var seatIDNext int
	var nextPlayerMsg rcvMessage

	i := seatID

	for {
		i++
		if i == TABLE_PLAYERS_MAX {
			i = 0
		}

		if usersMsg[i].ConnType == "RESERVED" {
			continue
		} else if i == seatID || usersMsg[i].ConnType != "" {
			seatIDNext = i
			break
		}
	}

	// fmt.Println("seatIDNext", seatIDNext)
	nextPlayerMsg = usersMsg[seatIDNext]
	nextPlayerMsg.IsActivated = true
	if nextPlayerMsg.Status == "AUTO" {
		nextPlayerMsg.ConnType = "BNEXT"
		nextPlayerMsg.Greeting = "Hello"
		nextPlayerMsg.IsActivated = false
	}

	return nextPlayerMsg
}

func createAutoPlayersForSpecificTable(TableUsers [9]rcvMessage) [9]rcvMessage {

	var nickName = [...]string{"流逝的风", "每天赢5千", "不好就不要", "牛牛牛009", "风清猪的", "总是输没完了", "适度就是", "无畏了吗", "见好就收", "坚持到底", "三手要比", "不勉强", "搞不懂", "吴潇无暇", "大赌棍", "一直无感", "逍遥子", "风月浪", "独善其身", "赌神"}

	randomNums := generateRandomNumber(0, 19, 3)

	numofp := 0
	for i := 0; i < 9; i++ {
		if TableUsers[i].ConnType != "RESERVED" {
			numofp++
		}
	}

	newUserCounter := 0
	if numofp < 3 {
		for i := 0; i < 9; i++ {
			if TableUsers[i].ConnType == "RESERVED" {
				TableUsers[i].UserID = nickName[randomNums[newUserCounter]]
				TableUsers[i].Status = "AUTO"
				TableUsers[i].ConnType = "ASSINGED"
				TableUsers[i].IsActivated = false
				TableUsers[i].Round = 0
				TableUsers[i].SeatID = i
				TableUsers[i].Betvol = 100
				TableUsers[i].Greeting = "Helloo"
				newUserCounter++
				if newUserCounter > 2 {
					break
				}
			}
		}
	}
	return TableUsers
}

func tableInfoDevlivery(delay time.Duration, ch chan rcvMessage) {
	var t [VOL_TABLE_MAX]*time.Timer
	delayAuto := 6 * time.Second
	var TableUsers [VOL_TABLE_MAX][TABLE_PLAYERS_MAX]rcvMessage
	var nextPlayerMsg rcvMessage

	// TablesUsersMaps := make([]map[string]int, VOL_TABLE_MAX)
	for i := 0; i < VOL_TABLE_MAX; i++ {
		t[i] = time.NewTimer(delay)
		for j := 0; j < TABLE_PLAYERS_MAX; j++ {
			TableUsers[i][j].ConnType = "RESERVED"
		}
	}

	for {
		select {
		case rcv := <-ch:
			// save the users info of the specific table by seatID
			TableUsers[rcv.TableID][rcv.SeatID] = rcv
			fmt.Println(rcv, "--Received Msg")
			switch rcv.ConnType {
			case "JOINED":
				TableUsers[rcv.TableID] = createAutoPlayersForSpecificTable(TableUsers[rcv.TableID])
				sendChan <- TableUsers[rcv.TableID]
			case "BNEXT":
				nextPlayerMsg = createNextPlayerMsg(TableUsers[rcv.TableID], rcv.SeatID)
				TableUsers[nextPlayerMsg.TableID][nextPlayerMsg.SeatID] = nextPlayerMsg
			default:
				fmt.Println("Received ConnType:", rcv.ConnType)
			}

			fmt.Println(TableUsers)

			if nextPlayerMsg.Status == "AUTO" {
				t[rcv.TableID].Reset(delayAuto)
				// fmt.Println("Set dealy for ATUO", delayAuto)
			} else {
				t[rcv.TableID].Reset(delay)
			}
			continue
		case <-t[0].C:
			fmt.Println(nextPlayerMsg, "--T1 Timer trigger Send Msg-next player")
			//	if nextPlayerMsg.Status == "AUTO" {
			//		sendChan <- TableUsers[0]
			//	}
			t[0].Reset(delay)
		case <-t[1].C:
			fmt.Println("T2 no new player message, repeat time interval:", delay)
			// t[1].Reset(delay)
		case <-t[2].C:
			fmt.Println("T3 no new player message, repeat time interval:", delay)
			// t[2].Reset(delay)
		}
	}
}

func receiveJsonHandler(connection *websocket.Conn) {
	var rcv rcvMessage
	ch := make(chan rcvMessage)

	defer close(done)

	delay := 12 * time.Second
	go tableInfoDevlivery(delay, ch)

	for {
		err := connection.ReadJSON(&rcv)
		if err != nil {
			log.Println("Received not JSON data!")
			continue
		}
		ch <- rcv
	}
}

//生成count个[start,end)结束的不重复的随机数
func generateRandomNumber(start int, end int, count int) []int {
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}

	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn((end - start)) + start

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}
