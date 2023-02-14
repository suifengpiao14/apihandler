package onebehaviorentity

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/suifengpiao14/templatemap/util"
)

const (
	EOF = "\n"
)
const (
	SQL_TYPE_SELECT = "SELECT"
	SQL_TYPE_OTHER  = "OTHER"
	LOG_LEVEL_DEBUG = "debug"
	LOG_LEVEL_INFO  = "info"
	LOG_LEVEL_ERROR = "error"
)

type SQLActuatorI interface {
	GetSQL() (sql string, err error)
	GetDB() *sql.DB
}

type SQLActuator struct {
	SQLActuatorI
}

func NewSQLActuator(sqlActuatorI SQLActuatorI) (sqlActuator SQLActuatorI) {
	return SQLActuator{
		SQLActuatorI: sqlActuatorI,
	}
}

func (act SQLActuator) Exec(ctx context.Context, out interface{}) (err error) {
	sqls, err := act.GetSQL()
	if err != nil {
		return err
	}

	sqls = util.StandardizeSpaces(util.TrimSpaces(sqls)) // 格式化sql语句
	sqlType := SQLType(sqls)
	db := act.GetDB()
	if sqlType != SQL_TYPE_SELECT {
		_, err = db.Exec(sqls)
		//res, err := db.Exec(sqls)
		if err != nil {
			return err
		}
		// lastInsertId, _ := res.LastInsertId()
		// if lastInsertId > 0 {
		// 	return  nil
		// }
		// rowsAffected, _ := res.RowsAffected()
		return nil
	}
	rows, err := db.Query(sqls)
	if err != nil {
		return err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			panic(err)
		}
	}()
	allResult := make([][]map[string]string, 0)
	for {
		records := make([]map[string]string, 0)
		for rows.Next() {

			var record = make(map[string]interface{})
			var recordStr = make(map[string]string)
			err := MapScan(*rows, record)
			if err != nil {
				return err
			}
			for k, v := range record {
				if v == nil {
					recordStr[k] = ""
				} else {
					recordStr[k] = fmt.Sprintf("%s", v)
				}
			}
			records = append(records, recordStr)
		}
		allResult = append(allResult, records)
		if !rows.NextResultSet() {
			break
		}
	}

	if len(allResult) == 1 { // allResult 初始值为[[]],至少有一个元素
		result := allResult[0]
		if len(result) == 0 { // 结果为空，返回空字符串
			return nil
		}
		if len(result) == 1 && len(result[0]) == 1 {
			row := result[0]
			for _, val := range row {
				_ = val
				return nil // 只有一个值时，直接返回值本身
			}
		}
		return nil
	}
	return nil
}

//SQLType 判断 sql  属于那种类型
func SQLType(sqls string) string {
	sqlArr := strings.Split(sqls, EOF)
	selectLen := len(SQL_TYPE_SELECT)
	for _, sql := range sqlArr {
		if len(sql) < selectLen {
			continue
		}
		pre := sql[:selectLen]
		if strings.ToUpper(pre) == SQL_TYPE_SELECT {
			return SQL_TYPE_SELECT
		}
	}
	return SQL_TYPE_OTHER
}

//MapScan copy sqlx
func MapScan(r sql.Rows, dest map[string]interface{}) error {
	// ignore r.started, since we needn't use reflect for anything.
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	err = r.Scan(values...)
	if err != nil {
		return err
	}

	for i, column := range columns {
		dest[column] = *(values[i].(*interface{}))
	}

	return r.Err()
}
