package models

import (
	"database/sql"
	"fmt"
	"github.com/Qihoo360/doraemon/cmd/bash-command/initial"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/go-sql-driver/mysql"
	"strings"
	"sync"
)

var (
	globalOrm orm.Ormer
	once      sync.Once

	UserModel *userModel
)

func init() {
	// init orm tables
	orm.RegisterModel(
		new(User),
	)

	// init models
	UserModel = &userModel{}
}

const DbDriveName = "mysql"

func InitDb() {
	err := orm.RegisterDriver(DbDriveName, orm.DRMySQL)
	if err != nil {
		logs.Error("注册数据库驱动失败", err.Error())
		panic("注册数据库驱动失败")
	}

	err = ensureDB()
	if err != nil {
		logs.Error("确认数据库失败, ", err.Error())
		panic("确认数据库失败")
	}

}

func ensureDB() error {
	needInit := false
	dbName := beego.AppConfig.String("DBName")
	dbUrl := fmt.Sprintf("%s:%s@%s/", beego.AppConfig.String("DBUser"), beego.AppConfig.String("DBPasswd"), beego.AppConfig.String("DBTns"))
	fmt.Println(dbName, dbUrl)
	db, err := sql.Open(DbDriveName, fmt.Sprintf("%s%s", dbUrl, dbName))
	if err != nil {
		logs.Error("数据库登录失败,", dbName, err.Error())
		panic("数据库登录失败")
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		switch e := err.(type) {
		case *mysql.MySQLError:
			// MySQL error unkonw database;
			// refer https://dev.mysql.com/doc/refman/5.6/en/error-messages-server.html
			const MysqlErrNum = 1049
			if e.Number == MysqlErrNum {
				needInit = true
				dbForCreateDatabase, err := sql.Open(DbDriveName, addLocation(dbUrl))
				if err != nil {
					return err
				}
				defer dbForCreateDatabase.Close()
				_, err = dbForCreateDatabase.Exec(fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8 COLLATE utf8_general_ci;", dbName))
				if err != nil {
					return err
				}

			} else {
				return err
			}
		default:
			return err
		}
	}
	logs.Debug("Initialize database connection: %s", strings.Replace(dbUrl, beego.AppConfig.String("DBPasswd"), "****", 1))

	err = orm.RegisterDataBase("default", "mysql", addLocation(fmt.Sprintf("%s%s", dbUrl, dbName)))
	if err != nil {
		return err

	}

	if needInit {
		err = orm.RunSyncdb("default", false, true)
		if err != nil {
			return err
		}
		for _, insertSql := range initial.InitialData {
			_, err = orm.NewOrm().Raw(insertSql).Exec()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func addLocation(dbURL string) string {
	// https://stackoverflow.com/questions/30074492/what-is-the-difference-between-utf8mb4-and-utf8-charsets-in-mysql
	return fmt.Sprintf("%s?charset=utf8mb4&loc=%s", dbURL, beego.AppConfig.DefaultString("DBLoc", "Asia%2FShanghai"))
}

// singleton init ormer ,only use for normal db operation
// if you begin transaction，please use orm.NewOrm()
func Ormer() orm.Ormer {
	once.Do(func() {
		globalOrm = orm.NewOrm()
	})
	return globalOrm
}
