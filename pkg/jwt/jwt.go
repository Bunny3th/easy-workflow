package jwt

import (
	. "easy-workflow/pkg/config"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

//参考 https://github.com/golang-jwt/jwt

//生成自定义的Claims
func GenerateMyClaims(MyContent map[string]interface{}) jwt.MapClaims {
	//根据配置文件设定Token过期时间
	myClaims := jwt.MapClaims{"exp": time.Now().Add(time.Duration(JWT.ExpireDuration) * time.Second).Unix()}
	for k, v := range MyContent {
		myClaims[k] = v
	}
	return myClaims
}

//生成JWT Token
func GenToken(MyClaims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyClaims)
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(JWT.EncryptedString))
	return tokenString, err
}

//Token解析
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWT.EncryptedString), nil
	})
	//token在格式、编码校验、时间过期等情况下即会失效
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

//Token 续租
//考虑一种情况：假如用户访问过程中token失效，则会造成体验度降低
//所以可以在每次访问时，检验token是否将要过期，而后适当延长一些token过期时间
func TokenRenewal(Claims jwt.MapClaims) (string,error) {
	//获取传入Claims上的过期时间
	ExpirationTime,err:=Claims.GetExpirationTime()
	if err!=nil{
		return "",err
	}
	//在原有过期时间上增加续租时间
	expire := ExpirationTime.Add(time.Duration(JWT.RenewalDuration)*time.Second).Unix()
	//需要注意，Claims jwt.MapClaims本身是一个map。
	//而map作为函数参数，传递的是内存地址。所以必须在这里clone一个新的mao，才能不影响外部传入的初始map
	CloneClaims:=make(map[string]interface{})
	for k,v:=range Claims{
		CloneClaims[k]=v
	}

	CloneClaims["exp"] = expire
	if Token,err:=GenToken(CloneClaims);err==nil{
		return Token,nil
	}else{
		return "",err
	}
}
