package config

type Config_Gin struct {
	GinMode string
	Port    string
}

var Gin Config_Gin

func(i *Initializer) Gin() {
	Gin.GinMode = cfg.GetString("Gin.GinMode")
	Gin.Port = cfg.GetString("Gin.Port")
}
