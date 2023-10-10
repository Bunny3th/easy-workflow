package engine

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
)

//gorm参考文档 https://gorm.cn/zh_CN/docs/

var DB *gorm.DB

func DBConnect() error{
	//有关gorm.Config，可查看文档 https://gorm.cn/zh_CN/docs/gorm_config.html
	dsn := DBConnConfigurator.DBConnectString
	//gorm的默认日志是只打印错误和慢SQL,这里可以自定义日志级别
	myLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // （日志输出的目标，前缀和日志包含的内容）
		logger.Config{
			SlowThreshold:             time.Duration(DBConnConfigurator.SlowThreshold) * time.Second, // 慢SQL阈值
			LogLevel:                  logger.LogLevel(DBConnConfigurator.LogLevel),                  // 日志级别
			IgnoreRecordNotFoundError: DBConnConfigurator.IgnoreRecordNotFoundError,                  // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  DBConnConfigurator.Colorful,                                   // 使用彩色打印
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",   //表前缀.如前缀为t_，则`User` 的表名应该是 `t_users`
			SingularTable: true, //使用单数表名，启用该选项，此时，`User` 的表名应该是 `user`
		},
		Logger: myLogger,
	})
	if err!=nil{
		return err
	}

	//将局部变量db赋值给pkg变量DB
	//why?因为假设这么写 DB,err:= gorm.Open(),此时的DB只是一个新生成的局部变量，而非给全局变量DB赋值
	DB=db

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(DBConnConfigurator.MaxIdleConns)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(DBConnConfigurator.MaxOpenConns)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Minute * time.Duration(DBConnConfigurator.ConnMaxLifetime))

	return nil
}

/*
执行SQL语句，返回执行结果(可选)
示例：
ExecSQL("CALL SP_GET(?)",&RESULT,Par1,Par2)
ExecSQL("select * from test where id=? and c=?",&RESULT,1,"yes")
ExecSQL("update test set c="no" where id=? and c=?",nil,1,"yes")
*/
func ExecSQL(SQL string, Result interface{}, Params ...interface{}) (interface{}, error) {
	var d *gorm.DB

	//没有返回值，用db.Exec
	if Result == nil {
		if Params == nil { //无参数
			d = DB.Exec(SQL)
		} else { //有参数
			d = DB.Exec(SQL, Params...)
		}
		return "ok", d.Error
	}

	//有返回值，用db.Raw
	if Params == nil { //无参数
		d = DB.Raw(SQL).Scan(Result)
	} else { //有参数
		d = DB.Raw(SQL, Params...).Scan(Result)
	}
	return Result, d.Error
}
