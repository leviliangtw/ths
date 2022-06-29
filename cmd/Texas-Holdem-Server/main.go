package main

import (
	"crypto/md5"
	_ "crypto/md5"
	"database/sql"
	_ "encoding/xml"
	"fmt"
	"html/template"
	_ "html/template"
	"io"
	_ "io"
	"log"
	_ "math/rand"
	"net/http"
	"net/url"
	_ "net/url"
	"os"
	_ "os"
	"strconv"
	_ "strconv"
	"strings"
	"time"
	_ "time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"

	"github.com/leviliangtw/Texas-Holdem-Server/pkg/thslib"
)

func init() {
	//session, _ := store.Get(req, "PlayerInfo")

	/*
		store.Options = &sessions.Options{
			Domain:   "localhost",
			Path:     "/",
			MaxAge:   3600 * 1, // 1 hour
			HttpOnly: true,
		}
	*/
}

func GameStart(rw http.ResponseWriter, req *http.Request) {
	// allow cross domain AJAX requests
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	if req.Method == "POST" {
		req.ParseForm()

		db, err := sql.Open("mysql", "root:13014858@/texaspoker?charset=utf8") //連接資料庫
		thslib.CheckErr(err)

		//session試用，初始化session
		var store = sessions.NewCookieStore([]byte("texaspoker"))
		session, _ := store.Get(req, "PlayerInfo")
		session.Save(req, rw)

		if req.FormValue("SignIn") == "Sign In" { //非POST取值
			//SignIn(rw, req)
			thslib.SignIn(rw, req)
		} else if req.FormValue("SignOut") == "Sign Out" {
			thslib.SignOut(rw, req)
		} else if thslib.GameIsReady(db) { //判斷玩家是否全部登入(只要有玩家Role<0代表未取得ID)，為True代表五玩家均登入
			//判斷目前牌局階段，獲得Phrase
			Phrase := thslib.GetPhrase(db)
			if req.Form.Get("Info") == "GetTableSituation" { //玩家詢問牌桌情況
				var TTC thslib.TTCer //使用interface
				TTC = &thslib.TableToClient{}
				TTC.SendTableToClient(rw, db, Phrase)
			} else if req.Form.Get("Info") == "IsItMyTurn" { //玩家詢問輪到與否
				var TTS thslib.TTSer = thslib.GetTTS(req) //使用interface
				if TTS.IsPlayerWaited(db) {
					data := []byte("Info=YourTurn")
					affect, _ := rw.Write(data)
					fmt.Println("YourTurn: ", affect)
				} else {
					data := []byte("Info=NotYourTurn")
					affect, _ := rw.Write(data)
					fmt.Println("NotYourTurn: ", affect)
				}
			} else if req.FormValue("Shuffle") == "Shuffle" { //洗牌圈
				thslib.Shuffle(rw, req, db)
			} else if Phrase == 1 { //發手牌與翻牌前下注
				var TTS thslib.TTSer = thslib.GetTTS(req) //使用interface

				if TTS.IsPlayerWaited(db) {
					if TTS.IsDiscard() {
						TTS.UpdateDicardAction(rw, db)
						TTS.UpdateThePlayerwaitedAndPhrase(db, Phrase)
					} else {
						if TTS.IsEnoughChips(db) {
							TTS.UpdateActionAndChipInPot(db)
							TTS.UpdateThePlayerwaitedAndPhrase(db, Phrase)

							data := []byte("Info=YourTurnFinished")
							affect, _ := rw.Write(data)
							fmt.Println("YourTurnFinished: ", affect)

						} else {
							//POST ERR
							data := []byte("Info=NotEnoughChips")
							affect, _ := rw.Write(data)
							fmt.Println("Not Enough Chips: ", affect)
						}
					}
				} else {
					//POST ERR
					data := []byte("Info=WrongPlayer")
					affect, _ := rw.Write(data)
					fmt.Println("Wrong Player: ", affect)
				}
			} else if Phrase == 2 { //翻牌後
				var TTS thslib.TTSer = thslib.GetTTS(req) //使用interface
				if TTS.IsPlayerWaited(db) {
					if TTS.IsDiscard() {
						TTS.UpdateDicardAction(rw, db)
						TTS.UpdateThePlayerwaitedAndPhrase(db, Phrase)
					} else {
						if TTS.IsEnoughChips(db) {
							TTS.UpdateActionAndChipInPot(db)
							TTS.UpdateThePlayerwaitedAndPhrase(db, Phrase)

							data := []byte("Info=YourTurnFinished")
							affect, _ := rw.Write(data)
							fmt.Println("YourTurnFinished: ", affect)
						} else {
							//POST ERR
							data := []byte("Info=NotEnoughChips")
							affect, _ := rw.Write(data)
							fmt.Println("Not Enough Chips: ", affect)
						}
					}
				} else {
					//POST ERR
					data := []byte("Info=WrongPlayer")
					affect, _ := rw.Write(data)
					fmt.Println("WrongPlayer: ", affect)
				}
			} else if Phrase == 3 { //轉牌圈，最後即結束
				var TTS thslib.TTSer = thslib.GetTTS(req) //使用interface
				if TTS.IsPlayerWaited(db) {
					if TTS.IsDiscard() {
						TTS.UpdateDicardAction(rw, db)
						TTS.UpdateThePlayerwaitedAndPhrase(db, Phrase)
					} else {
						if TTS.IsEnoughChips(db) {
							TTS.UpdateActionAndChipInPot(db)
							TTS.UpdateThePlayerwaitedAndPhrase(db, Phrase)
							if thslib.IsFinalPhrase(db) { //最後一玩家下完注：
								var winner []int = thslib.WhoIsWinner(db) //判斷贏家
								thslib.CalculateTheChips(db, winner)      //結算籌碼
								thslib.ResetGame(db)                      //Inn加一，Role重設，win歸零，Action設-1，Phrase歸4(等於不更動)

								//POST "Winner"
								fmt.Println("Winner is: (0-4)", winner, "!!!")

								//POST "YourTurnFinished Final"
								data := []byte("Info=YourTurnFinished Final")
								affect, _ := rw.Write(data)
								fmt.Println("YourTurnFinished Final: ", affect)
							} else {
								//POST "YourTurnFinished"
								data := []byte("Info=YourTurnFinished")
								affect, _ := rw.Write(data)
								fmt.Println("YourTurnFinished: ", affect)
							}
						} else {
							//POST ERR
							data := []byte("Info=NotEnoughChips")
							affect, _ := rw.Write(data)
							fmt.Println("Not Enough Chips: ", affect)
						}
					}
				} else {
					//POST ERR
					data := []byte("Info=WrongPlayer")
					affect, _ := rw.Write(data)
					fmt.Println("WrongPlayer: ", affect)
				}
			}
		} else { //還有其他玩家未登入，而有玩家傳送洗牌或XML訊息過來
			//IP := req.RemoteAddr
			ID := req.Form.Get("team")
			//防止表單偽造，例如沒有第六組而有人傳送ID=6的封包
			slice := []string{"1", "2", "3", "4", "5"}
			for _, v := range slice {
				if v == ID {
					//BeforeGame_PostIdToClientAndSaveIp(db, IP, ID) //POST玩家ID至遠端玩家與儲存IP
					//thslib.BeforeGame_ModifyRoleByInn(db, ID) //獲得Inn以修改Role與IP
					//POST傳回ID
					PostData := []byte("ID=" + ID)
					affect, _ := rw.Write(PostData)
					fmt.Println("ID Post Back: ", affect)
				}
			}
		}
		fmt.Println(session.Values["ID"])
		db.Close()
	} else if req.Method == "GET" {
		var p thslib.HtmlItemer
		p = thslib.HtmlItem{}
		p.GETPageUpdate(rw, req)
	}
}

func main() {
	//設置訪問的路由
	http.HandleFunc("/", GameStart)

	http.HandleFunc("/SignIn", thslib.SignIn)
	http.HandleFunc("/SignOut", thslib.SignOut)
	http.HandleFunc("/sayhelloName", sayhelloName)
	http.HandleFunc("/login", login)
	http.HandleFunc("/upload", upload)

	//上傳檔案
	/*
		target_url := "http://localhost:80/upload"
		filename := "./astaxie.pdf"
		thslib.PostFile(filename, target_url)
	*/

	err := http.ListenAndServe(":8080", nil) //設置監聽的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", context.ClearHandler(http.DefaultServeMux))
	}
}

//Hello~~~先弄懂這裡！！！
func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "POST" {
		IP := r.RemoteAddr
		IP = "http://" + IP
		fmt.Println(IP)
		//fmt.Println("r.Form", r.Form)
		//fmt.Println("path", r.URL.Path)
		//fmt.Println("scheme", r.URL.Scheme)
		fmt.Println("From GameStart: ID=", r.Form["ID"])
		fmt.Println("From GameStart: Info=", r.Form["Info"])

		//data := make(url.Values)
		//add := "http://127.0.0.1:8000"
		//data.Set("Info", "It's Your Turn!!!")
		//resp, err := http.PostForm(IP, data)
		//resp, err := http.PostForm(IP, data)
		//checkErr(err)
		//defer resp.Body.Close()

		data := []byte("Info=It's Your Turn!!!")
		affect, _ := w.Write(data)
		fmt.Println("ResponseWriter length: ", affect)
	} else {
		//fmt.Println("sayhelloName: ", r.Form["ID"])
		IP := r.RemoteAddr
		IP = "http://" + IP
		//測試資料，通知指定玩家
		fmt.Println(IP)
		data := make(url.Values)
		add := "http://127.0.0.1:8000"
		data.Set("Info", "It's Your Turn!!!")
		resp, err := http.PostForm(add, data)
		//resp, err := http.PostForm(IP, data)
		thslib.CheckErr(err)
		defer resp.Body.Close()

		//fmt.Fprintln(w, "Hello astaxie!") //這個寫入到w的是輸出到客戶端的
	}
	fmt.Println(r.Form) //這些信息是輸出到服務器端的打印信息
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])

	v := r.Form
	v.Set("name", "Ava")
	v.Add("friend", "Jess")
	v.Add("friend", "Sarah")
	v.Add("friend", "Zoe")
	// v.Encode() == "name=Ava&friend=Jess&friend=Sarah&friend=Zoe"
	fmt.Println(v.Get("name"))
	fmt.Println(v.Get("friend"))
	fmt.Println(v["friend"])

	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}

}

//玩家也做Server端才用得到
func BeforeGame_PostIdToClientAndSaveIp(db *sql.DB, IP string, ID string) {
	//POST玩家ID至遠端玩家IP
	IP = "http://" + IP
	data := make(url.Values)

	//測試IP = "http://localhost:9090/sayhelloName"
	IP = "http://localhost:9090/sayhelloName"
	//測試IP = "http://localhost:9090/sayhelloName"

	data.Set("ID", ID)
	resp, err := http.PostForm(IP, data)
	thslib.CheckErr(err)
	defer resp.Body.Close()

	stmt, _ := db.Prepare("update playerinfo set IP=? where ID=?")
	res, _ := stmt.Exec(IP, ID)
	affect, _ := res.RowsAffected()
	fmt.Println("IP Update: ", affect)
}

func NotifyPlayerwaited(db *sql.DB, IdNext int) {
	//通知下一玩家
	data := make(url.Values)
	row := db.QueryRow("SELECT IP FROM PlayerInfo,TableSituation where PlayerWaited=IdNext")
	var IP string
	err := row.Scan(&IP)
	thslib.CheckErr(err)
	data.Set("Info", "It's Your Turn!!!")
	resp, err := http.PostForm(IP, data)
	thslib.CheckErr(err)
	defer resp.Body.Close()

	//測試資料，通知指定玩家
	add := "http://localhost:9090/sayhelloName"
	data.Set("Info", "It's Your Turn!!!")
	resp, err = http.PostForm(add, data)
	thslib.CheckErr(err)
	defer resp.Body.Close()
}

//未來有可能會用到，參考用
func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //獲取請求的方法
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("login.gtpl")
		t.Execute(w, token)
	} else {
		//請求的是登陸數據，那麼執行登陸的邏輯判斷
		r.ParseForm()

		token := r.Form.Get("token")
		if token != "" {
			//验证token的合法性
		} else {
			//不存在token报错
		}

		if len(r.Form.Get("username")) == 0 {
			//為空的處理
			fmt.Fprintf(w, "No UserName: %s", r.Form.Get("username"))
		} else {
			fmt.Fprintf(w, "UserName: %s", r.Form.Get("username"))
			fmt.Println("username:", r.Form["username"])
			fmt.Println("password:", r.Form["password"])

			slice := []string{"1", "2", "3", "4", "5"}
			for _, v := range slice {
				if v == r.Form.Get("team") {
					fmt.Println("team:", r.Form["team"])
				}
			}
			fmt.Fprintf(w, "NO!")

		}
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //獲取請求的方法
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("./upload/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

func Client(c *http.Client, data url.Values) {
	url := "http://localhost:9090/GameStart"
	data.Add("ID", "1050533008")
	c.PostForm(url, data)
}
