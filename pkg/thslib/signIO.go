package thslib

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("texaspoker"))

func SignIn(rw http.ResponseWriter, req *http.Request) {
	db, err := sql.Open("mysql", "root:13014858@/texaspoker?charset=utf8") //連接資料庫
	CheckErr(err)
	req.ParseForm()
	//判斷目前牌局階段，獲得Phrase
	row := db.QueryRow("SELECT Phrase FROM TableSituation")
	var Phrase int
	err = row.Scan(&Phrase)
	CheckErr(err)
	fmt.Println("Phrase Select: ", Phrase)

	//session試用，初始化session
	session, _ := store.Get(req, "PlayerInfo")
	if req.Form.Get("team") != "" { //
		session.Values["ID"] = req.Form.Get("team")
	}
	session.Save(req, rw)
	fmt.Println(session.Values["ID"])
	ID, OK := session.Values["ID"].(string)
	fmt.Println(OK)
	BeforeGame_ModifyRoleByInn(db, ID) //獲得Inn以修改Role與IP

	//顯示ID於網頁，表示已登入
	var p HtmlItemer
	p = HtmlItem{}
	p.SignInUpdate(rw, db, Phrase, ID)
}

func SignOut(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	db, err := sql.Open("mysql", "root:13014858@/texaspoker?charset=utf8") //連接資料庫
	CheckErr(err)

	session, _ := store.Get(req, "PlayerInfo")
	ID, OK := session.Values["ID"].(string)

	BeforeGame_ResetRoleByInn(db, ID)

	fmt.Println(OK)
	session.Options = &sessions.Options{MaxAge: -1} //session.clear()
	session.Save(req, rw)

	//不顯示ID於網頁，表示已登出
	var p HtmlItemer
	p = HtmlItem{}
	p.SignOutUpdate(rw, ID)
}

func BeforeGame_ModifyRoleByInn(db *sql.DB, ID string) {
	//獲得Inn以修改Role
	row := db.QueryRow("SELECT ID, Inn FROM playerinfo where ID=?", ID)
	var i_ID, Inn, Role int
	err := row.Scan(&i_ID, &Inn)
	CheckErr(err)
	for i_ID-Inn <= 0 {
		i_ID += 5
	}
	Role = i_ID - Inn
	stmt, _ := db.Prepare("update playerinfo set Role=? where ID=?")
	res, _ := stmt.Exec(Role, ID)
	affect, _ := res.RowsAffected()
	fmt.Println("Role and IP Update: ", affect)
}

func BeforeGame_ResetRoleByInn(db *sql.DB, ID string) {
	stmt, _ := db.Prepare("update playerinfo set Role=-1 where ID=?")
	res, _ := stmt.Exec(ID)
	affect, _ := res.RowsAffected()
	fmt.Println("Role and IP Update: ", affect)
}
