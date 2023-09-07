package config

type DataBaseConnect struct {
	DBConnectString string //连接字符串
	MaxIdleConns    int    //空闲连接池中连接的最大数量
	MaxOpenConns    int    //打开数据库连接的最大数量
	ConnMaxLifetime int    //连接可复用的最大时间（分钟）
}

type DataBaselog struct {
	SlowThreshold             int64 //慢SQL阈值(秒)
	LogLevel                  int   //日志级别 1:Silent  2:Error 3:Warn 4:Info
	IgnoreRecordNotFoundError bool  //忽略ErrRecordNotFound（记录未找到）错误
	Colorful                  bool  //使用彩色打印
}

var DBConnect = DataBaseConnect{MaxIdleConns: 10, MaxOpenConns: 100, ConnMaxLifetime: 3600}
var DBlog = DataBaselog{SlowThreshold: 1, LogLevel: 4, IgnoreRecordNotFoundError: true, Colorful: true}


