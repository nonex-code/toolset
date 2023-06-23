package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type ConnConfig struct {
	*sqlx.DB
}

func NewDBConnPool(driverName, dataSource string) (*ConnConfig, error) {
	db, err := sqlx.Open(driverName, dataSource)
	if err != nil {
		return nil, err
	}
	//// 设置连接池中空闲连接的最大数量。
	//db.SetMaxIdleConns(10)
	//// 设置打开数据库连接的最大数量。
	//db.SetMaxOpenConns(100)
	//// 设置连接可复用的最大时间。
	//db.SetConnMaxLifetime(time.Second * 30)
	////设置连接最大空闲时间
	//db.SetConnMaxIdleTime(time.Second * 30)
	//检查连通性
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &ConnConfig{db}, nil
}

func (m *ConnConfig) QueryDataToMap(queryStr string, args ...interface{}) ([]map[string]string, int64, error) {

	rows, err := m.Query(queryStr, args...)
	defer rows.Close()
	if err != nil {
		return nil, 0, err
	}
	//获取列名cols
	cols, _ := rows.Columns()
	if len(cols) > 0 {
		var ret []map[string]string
		for rows.Next() {
			buff := make([]interface{}, len(cols))
			data := make([][]byte, len(cols)) //数据库中的NULL值可以扫描到字节中
			for i, _ := range buff {
				buff[i] = &data[i]
			}
			rows.Scan(buff...) //扫描到buff接口中，实际是字符串类型data中
			//将每一行数据存放到数组中
			dataKv := make(map[string]string, len(cols))
			for k, col := range data { //k是index，col是对应的值
				dataKv[cols[k]] = string(col)
			}
			ret = append(ret, dataKv)
		}

		return ret, int64(len(ret)), nil
	} else {
		return nil, 0, err
	}
}

func (m *ConnConfig) UIData(sql string, args ...interface{}) (int64, error) {
	exec, err := m.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	Affected, _ := exec.RowsAffected()
	//log.Println("Affected", affected)
	return Affected, nil
}
func (m *ConnConfig) QueryDataToStruct(obj any, queryStr string, args ...interface{}) error {
	err := m.Select(obj, queryStr, args...)
	if err != nil {
		return err
	}
	return err
}
