package config

type Config_DataBaseConnect struct {
	DBConnectString string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

type Config_DataBaselog struct {
	SlowThreshold             int64
	LogLevel                  int
	IgnoreRecordNotFoundError bool
	Colorful                  bool
}

type Config_DataBaseNamingStrategy struct {
	TablePrefix   string
	SingularTable bool
}

var DBConnect Config_DataBaseConnect
var DBlog Config_DataBaselog
var DBNamingStrategy Config_DataBaseNamingStrategy

func (i *Initializer) DataBase() {
	DBConnect.DBConnectString = cfg.GetString("DataBase.Connect.DBConnectString")
	DBConnect.MaxIdleConns = cfg.GetInt("DataBase.Connect.MaxIdleConns")
	DBConnect.MaxOpenConns = cfg.GetInt("DataBase.Connect.MaxOpenConns")
	DBConnect.ConnMaxLifetime = cfg.GetInt("DataBase.Connect.ConnMaxLifetime")

	DBlog.SlowThreshold=cfg.GetInt64("DataBase.log.SlowThreshold")
	DBlog.LogLevel= cfg.GetInt("DataBase.log.LogLevel")
	DBlog.IgnoreRecordNotFoundError=cfg.GetBool("DataBase.log.IgnoreRecordNotFoundError")
	DBlog.Colorful=cfg.GetBool("DataBase.log.Colorful")

	DBNamingStrategy.TablePrefix=cfg.GetString("DataBase.NamingStrategy.TablePrefix")
	DBNamingStrategy.SingularTable=cfg.GetBool("DataBase.NamingStrategy.SingularTable")
}
