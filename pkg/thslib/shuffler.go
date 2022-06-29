package thslib

import (
	"database/sql"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

func Shuffle(rw http.ResponseWriter, req *http.Request, db *sql.DB) {

	stmt, _ := db.Prepare("UPDATE TableSituation SET Phrase=1 WHERE Phrase=4")
	res, _ := stmt.Exec()
	affect, _ := res.RowsAffected()
	fmt.Println("TableSituation Phrase Reset!!!", affect)

	var store = sessions.NewCookieStore([]byte("texaspoker"))
	session, _ := store.Get(req, "PlayerInfo")
	ID, OK := session.Values["ID"].(string)
	fmt.Println(OK)
	//遊戲最初洗牌與UPDATE(0-51)
	a := Deal()
	UpdateCards(db, a)
	UpdateBlind(db)
	UpdatePlayerwaited_Beginning(db)
	//UpdatePhrase_Beginning(db, Phrase)

	t, _ := template.ParseFiles("login.gtpl")

	p := HtmlItem{
		Token:            "",
		ID:               ID,
		FieldsetDisable:  "disabled",
		IsDisplay:        "true",
		SignInDisplay:    "none",
		SignOutDisplay:   "true",
		StartGameDisplay: "none",
		ExitGameDisplay:  "none"}
	t.Execute(rw, p)
}

func Deal() []int {
	a := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,
		26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
		39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51}
	rand.Seed(time.Now().UnixNano())
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
	return a
}

func UpdateCards(db *sql.DB, a []int) {
	//Update公共牌
	stmt, _ := db.Prepare("UPDATE TableSituation SET CardF1=?, CardF2=?, CardF3=?, CardT=?, CardR=?")
	res, _ := stmt.Exec(a[0], a[1], a[2], a[3], a[4])
	affect, _ := res.RowsAffected()
	fmt.Println("Public Deal Update: ", affect)
	//Update各玩家手牌
	for i := 1; i <= 5; i++ {
		stmt, _ := db.Prepare("UPDATE playerinfo SET Win=0,Card1=?, Card2=? where ID=?")
		res, _ := stmt.Exec(a[2*i+3], a[2*i+4], i)
		affect, _ := res.RowsAffected()
		fmt.Println("Private Deal Update: ", affect)
	}
}

func UpdateBlind(db *sql.DB) {
	//Update盲注
	for i := 1; i <= 2; i++ {
		stmt, _ := db.Prepare("UPDATE playerinfo SET Chips=Chips-?, ChipsInPot=? WHERE Role=?")
		res, _ := stmt.Exec(25*i, 25*i, i)
		affect, _ := res.RowsAffected()
		fmt.Println("Blind Update: ", affect)
	}
}

func UpdatePlayerwaited_Beginning(db *sql.DB) {
	//Update等待玩家與牌局階段
	row := db.QueryRow("SELECT ID From playerinfo WHERE Role=3")
	var ID int
	err := row.Scan(&ID)
	CheckErr(err)
	stmt, _ := db.Prepare("UPDATE TableSituation SET PlayerWaited=?")
	res, _ := stmt.Exec(ID)
	affect, _ := res.RowsAffected()
	fmt.Println("PlayerWaited Update: ", ID, affect)
}
