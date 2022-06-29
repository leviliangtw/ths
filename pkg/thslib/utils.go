package thslib

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func GameIsReady(db *sql.DB) bool {
	rows, _ := db.Query("SELECT Role FROM playerinfo")
	flag := true
	for rows.Next() {
		var Role int
		err := rows.Scan(&Role)
		CheckErr(err)
		if Role < 0 {
			flag = false
		}
	}
	return flag
}

func IsFinalPhrase(db *sql.DB) bool {
	var Phrase int
	row := db.QueryRow("SELECT Phrase FROM TableSituation")
	err := row.Scan(&Phrase)
	CheckErr(err)
	if Phrase == 4 {
		return true
	}
	return false
}

func CalculateTheChips(db *sql.DB, winner []int) {
	//結算籌碼
	var AllChipsInPot int = 0
	rows, _ := db.Query("SELECT ChipsInPot FROM PlayerInfo")
	for rows.Next() {
		var ChipsInPot int
		err := rows.Scan(&ChipsInPot)
		if err != nil {
			log.Fatal(err)
		}
		AllChipsInPot += ChipsInPot
	}
	for i := 0; i < len(winner); i++ {
		stmt, _ := db.Prepare("update playerinfo set Chips=Chips+?, ChipsInPot=0, Win=1 where ID=?")
		res, _ := stmt.Exec((AllChipsInPot / len(winner)), (winner[i] + 1))
		affect, _ := res.RowsAffected()
		fmt.Println("Final Chips Update!!!", affect)
	}
}

func ResetGame(db *sql.DB) {
	//Inn加一，Role重設，Action設-1，Phrase歸零
	var affect int64
	for i := 1; i <= 5; i++ {
		stmt, _ := db.Prepare("UPDATE PlayerInfo SET Inn=Inn+1, ChipsInPot=0, Action=-1 WHERE ID=?")
		res, _ := stmt.Exec(i)
		affect, _ = res.RowsAffected()

		//獲得Inn以修改Role
		row := db.QueryRow("SELECT ID, Inn FROM playerinfo where ID=?", i)
		var i_ID, Inn, Role int
		err := row.Scan(&i_ID, &Inn)
		CheckErr(err)
		for i_ID-Inn <= 0 {
			i_ID += 5
		}
		Role = i_ID - Inn
		stmt, _ = db.Prepare("update playerinfo set Role=? where ID=?")
		res, _ = stmt.Exec(Role, i)
		affect, _ = res.RowsAffected()
	}
	fmt.Println("Role and IP Update: ", affect)
	fmt.Println("PlayerInfo Reset!!!", affect)
	//stmt, _ := db.Prepare("UPDATE TableSituation SET Phrase=0 WHERE Phrase=4")
	//res, _ := stmt.Exec()
	//affect, _ = res.RowsAffected()
	//fmt.Println("TableSituation Reset!!!", affect)
}

func GetPhrase(db *sql.DB) int {
	//判斷目前牌局階段，獲得Phrase
	row := db.QueryRow("SELECT Phrase FROM TableSituation")
	var Phrase int
	err := row.Scan(&Phrase)
	CheckErr(err)
	fmt.Println("Phrase Select: ", Phrase)
	return Phrase
}

func PostFile(filename string, targetUrl string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//關鍵的一步操作
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	//打開文件句柄操作
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
