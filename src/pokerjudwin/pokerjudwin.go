// Texa'em Judwin
package pokerjudwin

import (
	"database/sql"
	"log"
)

func WhoIsWinner(db *sql.DB) []int {
	//fmt.Println("PASS1")
	//判斷贏家
	var winner []int
	var ontable [5]int
	var playerhd [5][2]int
	var playerstate [5]bool
	row := db.QueryRow("SELECT CardF1, CardF2, CardF3, CardT, CardR FROM TableSituation WHERE Phrase = 4")
	err := row.Scan(&ontable[0], &ontable[1], &ontable[2], &ontable[3], &ontable[4])
	//fmt.Println(ontable[0], ", ", ontable[1], ", ", ontable[2], ", ", ontable[3], ", ", ontable[4])
	checkErr(err)
	rows, _ := db.Query("SELECT Card1, Card2, Action FROM PlayerInfo")
	i := 0
	for rows.Next() {
		var Action [5]int
		err := rows.Scan(&playerhd[i][0], &playerhd[i][1], &Action[i])
		if err != nil {
			log.Fatal(err)
		}
		if Action[i] == 1 {
			playerstate[i] = true
		} else {
			playerstate[i] = false
		}
		i++
	}
	//fmt.Println(playerhd[0][0], ", ", playerhd[1][0], ", ", playerhd[2][0], ", ", playerhd[3][0], ", ", playerhd[4][0])
	//fmt.Println(playerstate[0], ", ", playerstate[1], ", ", playerstate[2], ", ", playerstate[3], ", ", playerstate[4])
	winner = Judwin(ontable, playerhd, playerstate)
	//fmt.Println(winner)
	return winner
}

func Judwin(ontable [5]int, playerhd [5][2]int, playerstate [5]bool) []int {
	var have [][4][13]bool //記錄牌面上玩家所有的牌
	having := [4][13]bool{{false, false, false, false, false, false, false, false, false, false, false, false, false}, {false, false, false, false, false, false, false, false, false, false, false, false, false}, {false, false, false, false, false, false, false, false, false, false, false, false, false}, {false, false, false, false, false, false, false, false, false, false, false, false, false}}
	//cardtype := [5]int{8, 8, 8, 8, 8}                                  //記錄五個人的牌型預設為highcard
	var cardrecord [5][]int //記錄牌型中的記錄
	var player []int        //記錄還在牌面上的玩家
	for i := 0; i < 5; i++ {
		if playerstate[i] == true {
			player = append(player, i)
		}
	}
	for i := 0; i < len(player); i++ {
		have = append(have, having)
	} //將各玩家的持有牌歸零
	for i := 0; i < len(player); i++ {
		for j := 0; j < 5; j++ {
			have[i][ontable[j]/13][ontable[j]%13] = true
		}
	} //把公牌丟入每個玩家的牌組記錄中
	for i := 0; i < len(player); i++ {
		for j := 0; j < 2; j++ {
			have[i][playerhd[player[i]][j]/13][playerhd[player[i]][j]%13] = true
		}
	} //將再檯面上的玩家手牌丟入
	for i := 0; i < len(player); i++ {
		if straightflush(have[i])[0] == 1 {
			straightfrecord := straightflush(have[i])
			straightfrecord[0] = 8
			cardrecord[i] = straightfrecord[0:2]
			continue
		} else if fourofkind(have[i])[0] == 1 {
			fourofkindrecord := fourofkind(have[i])
			fourofkindrecord[0] = 7
			cardrecord[i] = fourofkindrecord[0:3]
			continue
		} else if fullhouse(have[i])[0] == 1 {
			fullhouserecord := fullhouse(have[i])
			fullhouserecord[0] = 6
			cardrecord[i] = fullhouserecord[0:3]
			continue
		} else if flush(have[i])[0] == 1 {
			flushrecord := flush(have[i])
			flushrecord[0] = 5
			cardrecord[i] = flushrecord[0:6]
			continue
		} else if straight(have[0])[0] == 1 {
			straightrecord := straight(have[i])
			straightrecord[0] = 4
			cardrecord[i] = straightrecord[0:2]
			continue
		} else if threekind(have[0])[0] == 1 {
			threekindrecord := threekind(have[i])
			threekindrecord[0] = 3
			cardrecord[i] = threekindrecord[0:4]
			continue
		} else if twopair(have[i])[0] == 1 {
			twopairrecord := twopair(have[i])
			twopairrecord[0] = 2
			cardrecord[i] = twopairrecord[0:4]
			continue
		} else if pair(have[i])[0] == 1 {
			pairrecord := pair(have[i])
			pairrecord[0] = 1
			cardrecord[i] = pairrecord[0:5]
			continue
		} else {
			highcardrecord := [6]int{0, 0, 0, 0, 0, 0}
			temp := highcard(have[i])
			for i := 0; i < 5; i++ {
				highcardrecord[i+1] = temp[i]
			}
			cardrecord[i] = highcardrecord[0:6]
			continue
		}
	} //記錄每個玩家的牌型及關鍵點數至cardrecord中
	maxcardtype := 0
	for i := 0; i < len(player); i++ {
		if maxcardtype < cardrecord[i][0] {
			maxcardtype = cardrecord[i][0]
		}
	} //maxcardtype 記錄最大的牌型
	var win []int
	for i := 0; i < len(player); i++ {
		if cardrecord[i][0] == maxcardtype {
			win = append(win, i)
		}
	} //將有較大牌組的玩家加入win群組中

	for i := 1; i < len(cardrecord[win[0]]); i++ { //由cardrecord中的第二個數字開始比較
		maxcard := cardrecord[win[0]][i]
		var newin []int //記錄比較完後新的贏家
		for j := 0; j < len(win); j++ {
			if cardrecord[win[j]][i] > maxcard {
				maxcard = cardrecord[win[j]][i]
			}
		} //maxcard記錄第i張牌所有玩家中最大的牌
		for j := 0; j < len(win); j++ {
			if cardrecord[win[j]][i] == maxcard {
				newin = append(newin, j)
			}
		}
		win = newin
		if len(win) == 1 {
			break
		}
	}
	//判斷獲勝玩家部分程式
	var result []int
	for i := 0; i < len(win); i++ {
		result = append(result, player[win[i]])
	}
	return result
}

func straightflush(having [4][13]bool) [2]int {
	record := [2]int{0, 0}
	for i := 0; i < 4; i++ {
		if having[i][0] && having[i][9] && having[i][10] && having[i][11] && having[i][12] {
			record[0], record[1] = 1, 13
			return record
		}
		for j := 8; j >= 0; j-- {
			if having[i][j] && having[i][j+1] && having[i][j+2] && having[i][j+3] && having[i][j+4] {
				record[0], record[1] = 1, j+4
				return record
			}
		}
	}
	return record
} //用一個有兩個整數的矩陣 第一個數字為0,1 1表示該牌組為同花順，第二個數字記錄同花順最大的數字A~13 分別記錄為 0~12

func fourofkind(having [4][13]bool) [3]int {
	record := [3]int{0, 0, 0}
	for j := 0; j < 13; j++ {
		if having[0][j] && having[1][j] && having[2][j] && having[3][j] {
			having[0][j], having[1][j], having[2][j], having[3][j] = false, false, false, false
			if j == 0 {
				record[0], record[1] = 1, 13
			} else {
				record[0], record[1] = 1, j
			}
		}
	}
	if record[0] == 1 {
		for i := 0; i < 4; i++ {
			if having[i][0] {
				record[2] = 13
				return record
			}
		} //確認是否有A
		for j := 12; j >= 1; j-- {
			for i := 0; i < 4; i++ {
				if having[i][j] {
					record[2] = j
					return record
				}
			}
		} //記錄最大的數字
	}
	return record
} //用一個有三個整數的矩陣 第一個數字為0,1，第二個數字為鐵支的數字，第三個數字記錄其餘三張牌最大的那張牌

func fullhouse(having [4][13]bool) [3]int {
	record := [3]int{0, 0, 0}
	three, two := false, false
	threenum := 0
	twonum := 0
	howmany := 0
	for i := 0; i < 4; i++ {
		if having[i][0] {
			howmany++
		}
	}
	if howmany == 3 {
		three, threenum = true, 13
	} else if howmany == 2 {
		two, twonum = true, 13
	}
	for j := 12; j >= 1; j-- {
		howmany = 0
		for i := 0; i < 4; i++ {
			if having[i][j] {
				howmany++
			}
		}
		if howmany == 3 {
			if three {
				two, twonum = true, j
				goto final
			}
			three, threenum = true, j
		} else if howmany == 2 {
			if two == false {
				two, twonum = true, j
			}
		}
	}
final:
	if three && two {
		record[0], record[1], record[2] = 1, threenum, twonum
	}
	return record
} //判斷是否為葫蘆 若是則傳回一個三個數字的整數陣列，第二個數字為有三張牌的數字，第三個數字為有兩張牌的數字

func flush(having [4][13]bool) [6]int {
	record := [6]int{0, 0, 0, 0, 0, 0}
	howmany := 0
	color := 0
	for i := 0; i < 4; i++ {
		howmany = 0
		for j := 0; j < 13; j++ {
			if having[i][j] {
				howmany++
			}
		}
		if howmany >= 5 {
			color = i
			record[0] = 1
			break
		}
	}
	if record[0] == 1 {
		i := 1
		if having[color][0] {
			record[1], i = 13, 2
		} //判斷同花中有沒有A
		for j := 12; j >= 1; j-- {
			if having[color][j] && i <= 5 {
				record[i] = j
				i++
			}
		}
	}
	return record
} //判斷牌組是否有同花 並記錄同花中數字最大的五個數字

func straight(having [4][13]bool) [2]int {
	record := [2]int{0, 0}
	isit := [13]bool{false, false, false, false, false, false, false, false, false, false, false, false, false}
	for i := 0; i < 13; i++ {
		for j := 0; j < 4; j++ {
			if having[j][i] {
				isit[i] = true
			}
		}
	} //把有數字的數字填滿
	if isit[0] && isit[9] && isit[10] && isit[11] && isit[12] {
		record[0], record[1] = 1, 13
		return record
	}
	for i := 8; i >= 0; i-- {
		if isit[i] && isit[i+1] && isit[i+2] && isit[i+3] && isit[i+4] {
			record[0], record[1] = 1, i+4
			return record
		}
	}
	return record
} //判斷牌組是否有順子 並記錄順子中最大的數字
func threekind(having [4][13]bool) [4]int {
	record := [4]int{0, 0, 0, 0}
	howmany := 0
	other := [2]int{0, 0}
	for i := 0; i < 4; i++ {
		if having[i][0] {
			howmany++
		}
	}
	if howmany == 3 {
		having[0][0], having[1][0], having[2][0], having[3][0] = false, false, false, false
		record[0], record[1] = 1, 13
		goto fdothers
	}
	for j := 12; j >= 1; j-- {
		howmany = 0
		for i := 0; i < 4; i++ {
			if having[i][j] {
				howmany++
			}
		}
		if howmany == 3 {
			having[0][j], having[1][j], having[2][j], having[3][j] = false, false, false, false
			record[0], record[1] = 1, j
			goto fdothers
		}
	}
fdothers: //判斷其他兩張牌的數字
	others := 0
	if record[0] == 1 {
		for i := 0; i < 4; i++ {
			if having[i][0] {
				other[others] = 13
				others++
				break
			}
		}
		for j := 12; j >= 1; j-- {
			for i := 0; i < 4; i++ {
				if having[i][j] {
					other[others] = j
					others++
					break
				}
			}
			if others == 2 {
				break
			}
		}
		record[1], record[2] = other[0], other[1]
	}
	return record
} //判斷牌組是否有三條，並且記錄四個數字 第二個數字為三條的數字，第三個數字為最大的數字，第四個數字為次大的數字

func twopair(having [4][13]bool) [4]int {
	record := [4]int{0, 0, 0, 0}
	pair := [2]int{0, 0}
	pairs := 0   //記錄有幾個對子
	howmany := 0 //記錄該號碼有幾種花色
	for i := 0; i < 4; i++ {
		if having[i][0] {
			howmany++
		}
	}
	if howmany == 2 {
		pair[pairs] = 13
		having[0][0], having[1][0], having[2][0], having[3][0], pairs = false, false, false, false, pairs+1
	} //先確認A有無成對
	for j := 12; j >= 1; j-- {
		howmany = 0
		for i := 0; i < 4; i++ {
			if having[i][j] {
				howmany++
			}
			if howmany == 2 {
				pair[pairs] = j
				having[0][j], having[1][j], having[2][j], having[3][j], pairs = false, false, false, false, pairs+1 //將算入pair的牌組消除記錄
				if pairs == 2 {
					record[0], record[1], record[2] = 1, pair[0], pair[1] //有twopair 記錄牌組以及flag=1
					goto fdothers                                         //尋找另外一個最大的牌
				}
			}
		}
	}
fdothers:
	if pairs == 2 {
		for i := 0; i < 4; i++ {
			if having[i][0] {
				record[3] = 13
				return record
			}
		}
		for j := 12; j >= 1; j-- {
			for i := 0; i < 4; i++ {
				if having[i][j] {
					record[3] = j
					return record
				}
			}
		}
	} //找出剩餘的最大牌
	return record
} //判斷牌組是否有twopairs 有的話則傳回四個整數 第二個數字為最大的pair 第三個為次大的pair 第四個數字為剩餘三張牌中最大的數字

func pair(having [4][13]bool) [5]int {
	record := [5]int{0, 0, 0, 0, 0}
	howmany := 0
	for i := 0; i < 4; i++ {
		if having[i][0] {
			howmany++
		}
	}
	if howmany == 2 {
		record[0], record[1] = 1, 0
		having[0][0], having[1][0], having[2][0], having[3][0] = false, false, false, false
		goto fdothers
	}
	for j := 12; j >= 1; j-- {
		howmany = 0
		for i := 0; i < 4; i++ {
			if having[i][j] {
				howmany++
			}
		}
		if howmany == 2 {
			record[0], record[1] = 1, j
			having[0][j], having[1][j], having[2][j], having[3][j] = false, false, false, false
			goto fdothers
		}
	}

fdothers:
	if record[0] == 1 {
		others := [3]int{0, 0, 0}
		othersnum := 0
		for i := 0; i < 4; i++ {
			if having[i][0] {
				others[othersnum] = 0
				othersnum++
				break
			}
		}
		for j := 12; j >= 0; j-- {
			for i := 0; i < 4; i++ {
				if having[i][j] {
					others[othersnum] = j
					othersnum++
					break
				}
			}
			if othersnum == 3 {
				break
			}
		}
		record[2], record[3], record[4] = others[0], others[1], others[2]
	}
	return record
} //判斷牌組是否為pair 傳回五個數字 第二個數字為pair的數字 其餘由大至小選取

func highcard(having [4][13]bool) [5]int {
	record := [5]int{0, 0, 0, 0, 0}
	recordsnum := 0
	for i := 0; i < 4; i++ {
		if having[i][0] {
			record[recordsnum] = 13
			recordsnum++
			break
		}
	}
	for j := 12; j >= 0; j-- {
		for i := 0; i < 4; i++ {
			if having[i][j] {
				record[recordsnum] = j
				recordsnum++
				break
			}
		}
		if recordsnum == 5 {
			break
		}
	}
	return record
} //記錄高牌中由大到小的五張牌

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
