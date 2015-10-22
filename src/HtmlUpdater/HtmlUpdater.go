package HtmlUpdater

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"

	"Unclassified"
)

type HtmlItem struct {
	Token            string
	ID               string
	FieldsetDisable  string
	IsDisplay        string
	SignInDisplay    string
	SignOutDisplay   string
	StartGameDisplay string
	ExitGameDisplay  string
}

type HtmlItemer interface {
	SignInUpdate(rw http.ResponseWriter, db *sql.DB, Phrase int, ID string)
	SignOutUpdate(rw http.ResponseWriter, ID string)
	ShuffleUpdate(rw http.ResponseWriter, ID string)
	GETPageUpdate(rw http.ResponseWriter, req *http.Request)
}

func (HtmIt HtmlItem) SignInUpdate(rw http.ResponseWriter, db *sql.DB, Phrase int, ID string) {
	//顯示ID於網頁，表示已登入
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
	if Unclassified.GameIsReady(db) { //判斷玩家是否全部登入(只要有玩家Role<0代表未取得ID)
		p.StartGameDisplay = "true"
		p.ExitGameDisplay = "true"
	}
	if Phrase >= 1 {
		p.StartGameDisplay = "none"
	}
	if Phrase == 4 {
		p.StartGameDisplay = "true"
	}

	t.Execute(rw, p)
}
func (HtmIt HtmlItem) SignOutUpdate(rw http.ResponseWriter, ID string) {
	//不顯示ID於網頁，表示已登出
	t, _ := template.ParseFiles("login.gtpl")
	p := HtmlItem{
		Token:            "",
		ID:               ID,
		FieldsetDisable:  "",
		IsDisplay:        "none",
		SignInDisplay:    "true",
		SignOutDisplay:   "none",
		StartGameDisplay: "none",
		ExitGameDisplay:  "none"}
	t.Execute(rw, p)
}
func (HtmIt HtmlItem) ShuffleUpdate(rw http.ResponseWriter, ID string) {
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
func (HtmIt HtmlItem) GETPageUpdate(rw http.ResponseWriter, req *http.Request) {
	db, err := sql.Open("mysql", "root:13014858@/texaspoker?charset=utf8") //連接資料庫
	fmt.Println(err)
	Unclassified.CheckErr(err)
	//判斷目前牌局階段，獲得Phrase
	row := db.QueryRow("SELECT Phrase FROM TableSituation")
	var Phrase int
	err = row.Scan(&Phrase)
	Unclassified.CheckErr(err)
	fmt.Println("Phrase Select: ", Phrase)
	//網頁登入頁面，獲得ID並請求開局
	crutime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(crutime, 10))
	md5_code := fmt.Sprintf("%x", h.Sum(nil))
	t, _ := template.ParseFiles("login.gtpl")

	var store = sessions.NewCookieStore([]byte("texaspoker"))
	session, _ := store.Get(req, "PlayerInfo")
	ID, OK := session.Values["ID"].(string)
	p := HtmlItem{
		Token:            md5_code,
		ID:               "0",
		FieldsetDisable:  "",
		IsDisplay:        "none",
		SignInDisplay:    "true",
		SignOutDisplay:   "none",
		StartGameDisplay: "none",
		ExitGameDisplay:  "none"}
	if OK != false {
		p.ID = ID
		p.FieldsetDisable = "disabled"
		p.IsDisplay = "true"
		p.SignInDisplay = "none"
		p.SignOutDisplay = "true"
	}
	if Unclassified.GameIsReady(db) { //判斷玩家是否全部登入(只要有玩家Role<0代表未取得ID)
		p.StartGameDisplay = "true"
		p.ExitGameDisplay = "true"
	}
	if Phrase >= 1 {
		p.StartGameDisplay = "none"
	}
	if Phrase == 4 {
		p.StartGameDisplay = "true"
	}
	t.Execute(rw, p)
}
