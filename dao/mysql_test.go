package dao

import (
	"fmt"
	"testing"
)

// 测试连接数据库，同时创建数据表
func Test_NewMqSQLConnector(t *testing.T) {
	option := MysqlOptions{
		Hostname:           "127.0.0.1", // 本地数据库
		Port:               "3306",      // 默认端口
		DbName:             "btc_relay", // 比特币的遍历器数据库名称
		User:               "root",      // 用户名
		Password:           "123aaa",    // 密码
		TablePrefix:        "btc_",      // 表格前缀
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    15,
	}
	tables := []interface{}{}
	tables = append(tables, Block{}, Transaction{}) // 添加表格数据结构体
	NewMqSQLConnector(&option, tables)              // 传参进去，对应的结构体将会被 xorm 自动解析并创建表
	fmt.Println("创建表格成功")
}
