// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package idb

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/idcsource/insight00-lib/cpool"
	"github.com/idcsource/insight00-lib/ilogs"
)

type DB struct {
	conf *cpool.Section
	*sql.DB
	Type int
	logs *ilogs.Logs
}

// 新建数据库连接。
// 此函数根据传入配置文件中关于数据库连接的信息进行配置，使用Golang自己的数据库接口，返回*sql.DB类型的数据库连接。
// 目前只处理针对PostgreSQL数据库的连接。
func NewDatabase(conf *cpool.Section, logs *ilogs.Logs) (db *DB, err error) {
	db_type, e := conf.GetConfig("type")
	if e != nil {
		err = errors.New("The configure have no database.type")
		if logs != nil {
			logs.ErrLog(err)
		}
		return
	}
	db = &DB{conf: conf, logs: logs}
	switch db_type {
	case "postgres":
		db.Type = DATABASE_TYPE_POSTGRESQL
		err = db.connPostgres()
		if err != nil {
			if logs != nil {
				logs.ErrLog(err)
			}
			return
		}
	default:
		err = errors.New("Can't use this database type : " + db_type)
		if logs != nil {
			logs.ErrLog(err)
		}
		return
	}
	return
}

func (d *DB) connPostgres() (err error) {
	db_server, e1 := d.conf.GetConfig("server")
	db_port, e2 := d.conf.GetConfig("port")
	db_user, e3 := d.conf.GetConfig("user")
	db_passwd, e4 := d.conf.GetConfig("passwd")
	db_dbname, e5 := d.conf.GetConfig("dbname")
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
		err = errors.New("Configuration is incomplete !")
		if d.logs != nil {
			d.logs.ErrLog(err)
		}
		return
	}
	var errs error
	connection_string := fmt.Sprintf("dbname=%s user=%s password=%s host=%s port=%s sslmode=disable", db_dbname, db_user, db_passwd, db_server, db_port)
	d.DB, errs = sql.Open("postgres", connection_string)
	if errs != nil {
		err := errors.New("Can't connect database : " + errs.Error())
		if d.logs != nil {
			d.logs.ErrLog(err)
		}
		return err
	}
	return nil
}
