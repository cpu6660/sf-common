package db

import (
	"fmt"
	"github.com/cpu6660/sf-common/conf"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"sync"
	"time"
)

const (
	DB_CONNECT_MODE_NEW = iota //每次重新创建连接
	DB_CONNECT_MODE_GET        //如果连接存在,复用
)

const (
	MaxTryCount     = 3   //SQL最大尝试连接数
	TryTimeInterval = 3   //Retry Time Interval
	MaxIdleConns    = 100 //MaxIdleConns
	MaxOpenConns    = 500 //MaxOpenConns
)

var DbClientsSingle *DbClients
var dbMutex sync.Mutex

type DbClients struct {
	clients map[string]*gorm.DB
	config  *conf.Config
	sync.Mutex
}

func NewDbClients(config *conf.Config, single bool) *DbClients {

	if single && DbClientsSingle != nil {
		return DbClientsSingle
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	dbClients := &DbClients{}
	dbClients.clients = make(map[string]*gorm.DB)
	dbClients.config = config

	if single {
		DbClientsSingle = dbClients
	}
	return dbClients
}

//get db conn
func (dbClients *DbClients) GetConn(dbName string, connectMode int) (*gorm.DB, error) {

	var (
		conn *gorm.DB
		err  error
	)

	//lock db clients  if create new  db client
	dbClients.Lock()

	if connectMode == DB_CONNECT_MODE_GET {
		if currentDb, ok := dbClients.clients[dbName]; ok {
			//check current client is connect or not
			err = currentDb.DB().Ping()
			if err == nil {
				return dbClients.clients[dbName], nil
			}
		}
	}

	defer dbClients.Unlock()

	//if client is disconnected, delete it
	if err != nil {
		//del db_name
		delete(dbClients.clients, dbName)
	}

	driver := dbClients.config.GetString(dbName + ":driver")
	userName := dbClients.config.GetString(dbName + ":user_name")
	password := dbClients.config.GetString(dbName + ":password")
	host := dbClients.config.GetString(dbName + ":host")
	db := dbClients.config.GetString(dbName + ":db_name")
	option := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local", userName, password, host, db)

	tryCount := MaxTryCount
	for conn, err = gorm.Open(driver, option); (err != nil) && (tryCount > 0); {
		tryCount--
		time.Sleep(TryTimeInterval * time.Second)
		conn, err = gorm.Open(driver, option)
	}

	if err != nil {
		return nil, err
	}

	if connectMode == DB_CONNECT_MODE_GET {
		dbClients.clients[dbName] = conn
	}

	conn.DB().SetMaxOpenConns(MaxOpenConns)
	conn.DB().SetMaxIdleConns(MaxIdleConns)
	conn.DB().SetConnMaxLifetime(1 * time.Hour)

	return conn, nil
}
