package XmlHandler

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"

	"Unclassified"
)

type TableToServer struct {
	XMLName    xml.Name `xml:"Player"`
	Id         int      `xml:"ID"`
	Action     int      `xml:"Action"`
	ChipsAdded int      `xml:"ChipsAdded"`
}
type TableToClient struct {
	XMLName xml.Name `xml:"Table"`
	Player  []Player `xml:"Player"`
	Public  []Public `xml:"Public"`
}
type Player struct {
	Id         int `xml:"ID"`
	Chips      int `xml:"Chips"`
	Inn        int `xml:"Inn"`
	Role       int `xml:"Role"`
	Chipsinpot int `xml:"ChipsInPot"`
	Win        int `xml:"Win"`
	Card1      int `xml:"Card1"`
	Card2      int `xml:"Card2"`
	Action     int `xml:"Action"`
}
type Public struct {
	Card_F1 int `xml:"Card_F1"`
	Card_F2 int `xml:"Card_F2"`
	Card_F3 int `xml:"Card_F3"`
	Card_T  int `xml:"Card_T"`
	Card_R  int `xml:"Card_R"`
}

type TTSer interface {
	IsEnoughChips(db *sql.DB) bool
	IsPlayerWaited(db *sql.DB) bool
	IsDiscard() bool
	UpdateDicardAction(rw http.ResponseWriter, db *sql.DB)
	UpdateActionAndChipInPot(db *sql.DB)
	UpdateThePlayerwaitedAndPhrase(db *sql.DB, Phrase int)
}

//public:
func (TTS TableToServer) IsEnoughChips(db *sql.DB) bool {
	//檢查是否有錯誤操作(未確實跟牌(看最高ChipsInPot))
	row := db.QueryRow("SELECT MAX(ChipsInPot) FROM PlayerInfo")
	var ChipsInPotBefore int
	err := row.Scan(&ChipsInPotBefore)
	checkErr(err)
	row = db.QueryRow("SELECT Chips, ChipsInPot FROM PlayerInfo where ID=?", TTS.Id)
	var Chips, ChipsInPotNow int
	err = row.Scan(&Chips, &ChipsInPotNow)
	checkErr(err)
	if TTS.ChipsAdded == Chips {
		return true
	} else if TTS.ChipsAdded+ChipsInPotNow < ChipsInPotBefore {
		return false
	}
	return true
}
func (TTS TableToServer) IsPlayerWaited(db *sql.DB) bool {
	//檢查玩家是否為PlayerWaited
	Id := TTS.Id
	row := db.QueryRow("SELECT PlayerWaited FROM TableSituation ")
	var PlayerWaited int
	err := row.Scan(&PlayerWaited)
	checkErr(err)
	fmt.Println("TTS.Id: ", TTS.Id)
	fmt.Println("TTS.Action: ", TTS.Action)
	fmt.Println("TTS.ChipsAdded: ", TTS.ChipsAdded)
	if Id == PlayerWaited {
		return true
	}
	return false
}
func (TTS TableToServer) IsDiscard() bool {
	if TTS.Action == 0 {
		return true
	} else {
		return false
	}
}
func (TTS TableToServer) UpdateDicardAction(rw http.ResponseWriter, db *sql.DB) {
	//更新Action
	stmt, _ := db.Prepare("update playerinfo set Action=0 where ID=?")
	res, _ := stmt.Exec(TTS.Id)
	affect1, _ := res.RowsAffected()
	fmt.Println("Action Update: ", affect1)
	//POST ERR
	data := []byte("Info=YouDiscard")
	affect, _ := rw.Write(data)
	fmt.Println("Info=YouDiscard: ", affect)
}
func (TTS TableToServer) UpdateActionAndChipInPot(db *sql.DB) {
	//更新Action與ChipInPot
	row := db.QueryRow("SELECT ChipsInPot FROM PlayerInfo where ID=?", TTS.Id)
	var ChipsInPotNow int
	err := row.Scan(&ChipsInPotNow)
	checkErr(err)
	stmt, _ := db.Prepare("update playerinfo set Action=1, Chips=Chips-?, ChipsInPot=? where ID=?")
	res, _ := stmt.Exec(TTS.ChipsAdded, TTS.ChipsAdded+ChipsInPotNow, TTS.Id)
	affect, _ := res.RowsAffected()
	fmt.Println("Action and ChipInPot Update: ", affect)
}
func (TTS TableToServer) UpdateThePlayerwaitedAndPhrase(db *sql.DB, Phrase int) {
	//Update牌局階段準備
	var Role, ChangeRole int
	IdNext := TTS.NextPlayer(db)
	row := db.QueryRow("SELECT Role FROM PlayerInfo WHERE ID=?", TTS.Id)
	err := row.Scan(&Role)
	checkErr(err)

	rowtable := [6]bool{false, false, false, false, false, false}
	rows, _ := db.Query("SELECT Role FROM PlayerInfo Where Action <> 0")
	for rows.Next() {
		var role int
		err := rows.Scan(&role)
		checkErr(err)
		rowtable[role] = true
	}
	for i := 3; i <= 7; i++ {
		num := i % 5
		if num == 0 {
			num = 5
		}
		if rowtable[num] {
			ChangeRole = num
		}
	}

	fmt.Println("ChangeRole", ChangeRole)
	fmt.Println("rowtable", rowtable)

	if Role == ChangeRole {
		stmt, _ := db.Prepare("update TableSituation set Phrase=? where Phrase=?")
		res, _ := stmt.Exec(Phrase+1, Phrase)
		affect, _ := res.RowsAffected()
		fmt.Println("Phrase Update: ", affect)
		//Update等待玩家
		stmt, _ = db.Prepare("update TableSituation set PlayerWaited=? where Phrase=?")
		res, _ = stmt.Exec(IdNext, Phrase+1)
		affect, _ = res.RowsAffected()
		fmt.Println("PlayerWaited Update: ", affect)
	} else {
		//Update等待玩家
		stmt, _ := db.Prepare("update TableSituation set PlayerWaited=? where Phrase=?")
		res, _ := stmt.Exec(IdNext, Phrase)
		affect, _ := res.RowsAffected()
		fmt.Println("PlayerWaited Update: ", affect)
	}
}

func GetTTS(req *http.Request) TableToServer {
	var TTS TableToServer //使用interface
	XmlFromClient := req.Form.Get("XmlToServer")
	fmt.Println("XmlFromClient: ", XmlFromClient)
	err := xml.Unmarshal([]byte(XmlFromClient), &TTS) //讀取玩家POST來的XML string
	Unclassified.CheckErr(err)
	return TTS
}

//private:
func (TTS TableToServer) NextPlayer(db *sql.DB) int {
	var IdNextFlag bool = true
	IdNext := TTS.Id
	for IdNextFlag {
		var Action int
		IdNext = IdNext + 1
		if IdNext == 6 {
			IdNext = 1
		}
		row := db.QueryRow("SELECT Action FROM PlayerInfo WHERE ID=?", IdNext)
		err := row.Scan(&Action)
		checkErr(err)
		if Action != 0 {
			IdNextFlag = false
		}
	}
	return IdNext
}

type TTCer interface {
	SendTableToClient(rw http.ResponseWriter, db *sql.DB, Phrase int)
}

//public:
func (TTC *TableToClient) SendTableToClient(rw http.ResponseWriter, db *sql.DB, Phrase int) {
	TTC.PrepareTTCforPublicCard(db, Phrase)
	//TTC發送
	output, err := xml.MarshalIndent(TTC, "  ", "    ")
	Unclassified.CheckErr(err)
	XmlToClient := xml.Header
	XmlToClient += string(output)
	//POST傳回TTC
	PostData := []byte("XmlToClient=" + XmlToClient)
	affect, _ := rw.Write(PostData)
	fmt.Println("TTC Post Out: ", affect)
}

//private:
func (TTC *TableToClient) PrepareTTCforPublicCard(db *sql.DB, Phrase int) {
	//TTC = &TableToClient{}
	rows, _ := db.Query("SELECT * FROM PlayerInfo")
	for rows.Next() {
		var (
			ID         int
			Chips      int
			Inn        int
			Role       int
			Chipsinpot int
			Win        int
			Card1      int
			Card2      int
			Action     int
			IP         string
		)
		err := rows.Scan(&ID, &Chips, &Inn, &Role, &Chipsinpot, &Win, &Card1, &Card2, &Action, &IP)
		if err != nil {
			log.Fatal(err)
		}
		TTC.Player = append(TTC.Player, Player{ID, Chips, Inn, Role, Chipsinpot, Win, Card1, Card2, Action})
	}
	//公共牌給各玩家(XML)準備
	if Phrase == 0 {
		TTC.Public = append(TTC.Public, Public{-1, -1, -1, -1, -1})
	} else if Phrase == 1 {
		TTC.Public = append(TTC.Public, Public{-1, -1, -1, -1, -1})
	} else if Phrase == 2 {
		rows, _ := db.Query("SELECT CardF1, CardF2, CardF3 FROM TableSituation")
		for rows.Next() {
			var (
				Card1 int
				Card2 int
				Card3 int
			)
			err := rows.Scan(&Card1, &Card2, &Card3)
			if err != nil {
				log.Fatal(err)
			}
			TTC.Public = append(TTC.Public, Public{Card1, Card2, Card3, -1, -1})
		}
	} else if Phrase == 3 {
		rows, _ := db.Query("SELECT CardF1, CardF2, CardF3, CardT FROM TableSituation")
		for rows.Next() {
			var (
				Card1 int
				Card2 int
				Card3 int
				CardT int
			)
			err := rows.Scan(&Card1, &Card2, &Card3, &CardT)
			if err != nil {
				log.Fatal(err)
			}
			TTC.Public = append(TTC.Public, Public{Card1, Card2, Card3, CardT, -1})
		}
	} else if Phrase == 4 {
		rows, _ := db.Query("SELECT CardF1, CardF2, CardF3, CardT, CardR FROM TableSituation")
		for rows.Next() {
			var (
				Card1 int
				Card2 int
				Card3 int
				CardT int
				CardR int
			)
			err := rows.Scan(&Card1, &Card2, &Card3, &CardT, &CardR)
			if err != nil {
				log.Fatal(err)
			}
			TTC.Public = append(TTC.Public, Public{Card1, Card2, Card3, CardT, CardR})
		}
	}
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
