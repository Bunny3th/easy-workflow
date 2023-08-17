package config

type DataBaseConnect struct {
	DBConnectString string
	MaxIdleConns    int //空闲连接池中连接的最大数量
	MaxOpenConns    int //打开数据库连接的最大数量
	ConnMaxLifetime int //连接可复用的最大时间（分钟）
}

type DataBaselog struct {
	SlowThreshold             int64 //慢SQL阈值(秒)
	LogLevel                  int   //日志级别 1:Silent  2:Error 3:Warn 4:Info
	IgnoreRecordNotFoundError bool  //忽略ErrRecordNotFound（记录未找到）错误
	Colorful                  bool  //使用彩色打印
}

var DBConnect = DataBaseConnect{MaxIdleConns: 10, MaxOpenConns: 100, ConnMaxLifetime: 3600}
var DBlog = DataBaselog{SlowThreshold: 1, LogLevel: 4, IgnoreRecordNotFoundError: true, Colorful: true}

//func (i *Initializer) DataBase() {
//	DBConnect.DBConnectString = cfg.GetString("DataBase.Connect.DBConnectString")
//	DBConnect.MaxIdleConns = cfg.GetInt("DataBase.Connect.MaxIdleConns")
//	DBConnect.MaxOpenConns = cfg.GetInt("DataBase.Connect.MaxOpenConns")
//	DBConnect.ConnMaxLifetime = cfg.GetInt("DataBase.Connect.ConnMaxLifetime")
//
//	DBlog.SlowThreshold=cfg.GetInt64("DataBase.log.SlowThreshold")
//	DBlog.LogLevel= cfg.GetInt("DataBase.log.LogLevel")
//	DBlog.IgnoreRecordNotFoundError=cfg.GetBool("DataBase.log.IgnoreRecordNotFoundError")
//	DBlog.Colorful=cfg.GetBool("DataBase.log.Colorful")
//
//	DBNamingStrategy.TablePrefix=cfg.GetString("DataBase.NamingStrategy.TablePrefix")
//	DBNamingStrategy.SingularTable=cfg.GetBool("DataBase.NamingStrategy.SingularTable")
//}


