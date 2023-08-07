package config

type Config_JWT struct {
	EncryptedString string
	ExpireDuration  int
	RenewalDuration int
}

var JWT Config_JWT

func(i *Initializer) JWT(){
	JWT.EncryptedString = cfg.GetString("JWT.EncryptedString")
	JWT.ExpireDuration = cfg.GetInt("JWT.ExpireDuration")
	JWT.RenewalDuration = cfg.GetInt("JWT.RenewalDuration")
}
