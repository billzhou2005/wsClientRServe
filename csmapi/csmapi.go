package csmapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Player struct {
	Name       string  `json:"name"`
	Cards      [3]Card `gorm:"embedded" json:"cards"`
	Cardstype  string  `json:"cardstype"`
	CIfirst    int     `json:"cifirst"`
	CIsecond   int     `json:"cisecond"`
	CIthird    int     `json:"cithird"`
	Cardsscore int     `json:"cardsscore"`
}

type Card struct {
	Points int `json:"points"`
	Suits  int `json:"suits"`
}

func GetPlayersCards(tableID int, numofp int) [9]Player {

	baseUrl := "http://140.143.149.188:8080/cardgen?"
	opt1 := "tableid="
	//	tableID := 50000011
	opt2 := "&numofp="
	//	numofp := 6

	url := baseUrl + opt1 + strconv.Itoa(tableID) + opt2 + strconv.Itoa(numofp)
	fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var players [9]Player
	json.Unmarshal(responseData, &players)

	return players
}
