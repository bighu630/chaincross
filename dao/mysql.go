package dao

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/ChainCross")
	if err != nil {
		log.Printf("File to Connect mysql %v", err)
		return
	}
	//设置数据库最大连接数
	DB.SetConnMaxLifetime(100)
	//设置上数据库最大闲置连接数
	DB.SetMaxIdleConns(10)
	//验证连接
	if err := DB.Ping(); err != nil {
		log.Println("open database fail ", err)
		return
	}
	log.Println("数据库连接成功")
}

// 查看已经存在的链名称
func CheckNameExist(names []string) []string {
	var existNames []string
	var nameList strings.Builder

	nameList.WriteString("(")
	for i, name := range names {
		nameList.WriteString("'")
		nameList.WriteString(name)
		nameList.WriteString("'")
		if i == len(names)-1 {
			break
		}
		nameList.WriteString(",")
	}
	nameList.WriteString(")")

	sql := "select * from ChainByName where name in " + nameList.String()
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("query incur error")
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			log.Printf("failt to sacn %v ", err)
		}
		existNames = append(existNames, name)
	}

	return existNames
}

// 通过chainID添加链
func AddChainsName(names []string) ([]string, error) {
	existName := CheckNameExist(names)
	if len(existName) != 0 {
		return existName, fmt.Errorf("some names has been used! %v ", existName)
	}

	//开启事务
	tx, err := DB.Begin()
	if err != nil {
		log.Println("tx fail")
		return nil, err
	}
	//准备sql语句
	for _, name := range names {
		stmt, err := tx.Prepare("INSERT INTO ChainByName (`name`) VALUES (?)")
		if err != nil {
			log.Println("Prepare fail")
			return nil, err
		}
		//将参数传递到sql语句中并且执行
		res, err := stmt.Exec(name)
		if err != nil {
			log.Println("Exec fail")
			return nil, err
		}
		log.Println(res.LastInsertId())

	}
	//将事务提交
	err = tx.Commit()
	//获得上一个插入自增的id
	return nil, err
}

func CreateHTCLTx(tx HTCLTx) (int64, error) {
	chainNames := []string{tx.ChainAName, tx.ChainBName}
	eixetChain := CheckNameExist(chainNames)
	if len(eixetChain) != 2 {
		return 0, fmt.Errorf("网关中仅包含%v", eixetChain)
	}
	t, err := DB.Begin()
	if err != nil {
		log.Println("tx fail")
		return 0, err
	}
	stmt, err := t.Prepare("INSERT INTO `ChainCross`.`htcl` (`ChainAName`,`TradeNFTID`, `NFTRecipientAddr`,`ChainBName`, `CoinNUM`, `CoinRecipientAddr`, `Hs`, `TimeStart`, `TimeEnd`,`AproveID`) VALUES(?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Println("Prepare fail")
		return 0, err
	}
	res, err := stmt.Exec(tx.ChainAName, tx.TradeNFTID, tx.NFTRecipientAddr, tx.ChainBName, tx.CoinNUM, tx.CoinRecipientAddr, tx.Hs, tx.TimeStart, tx.TimeEnd, tx.AproveID)
	if err != nil {
		log.Println("Exec fail")
		return 0, err
	}
	//将事务提交
	err = t.Commit()
	if err != nil {
		log.Println("数据库提交错误：", err)
		return 0, err
	}
	log.Println(res.LastInsertId())

	return res.LastInsertId()
}

func GetHTCLTx(id int64) (HTCLTx, error) {
	sqlStr := `SELECT * FROM htcl WHERE id=?`
	var tx HTCLTx
	var i int64
	err := DB.QueryRow(sqlStr, id).Scan(&i, &tx.ChainAName, &tx.TradeNFTID, &tx.NFTRecipientAddr, &tx.ChainBName, &tx.CoinNUM, &tx.CoinRecipientAddr, &tx.Hs, &tx.TimeStart, &tx.TimeEnd, &tx.AproveID)
	if err != nil {
		log.Printf("无法查找id为%d的htcl交易", id)
		return HTCLTx{}, err
	}
	return tx, nil
}

type HTCLTx struct {
	ChainAName        string
	TradeNFTID        string //交易NFTID
	NFTRecipientAddr  string // NFT接受者地址
	ChainBName        string
	CoinNUM           float64 //代币数量
	CoinRecipientAddr string  // 代币接受者地址
	Hs                string  //哈希时间锁用到的Hash(S)
	TimeStart         int64   // 开始时间戳
	TimeEnd           int64   //结束时间戳
	AproveID          string
}
