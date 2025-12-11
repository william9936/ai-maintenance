package tools

import (
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestGenToken(t *testing.T) {
	token, _ := GenToken(jwt.MapClaims{
		UserIdKey: strconv.Itoa(495511072137067261),
		RoleIdKey: 2,
		ExpKey:    time.Now().AddDate(0, 0, 30).Unix(), // 30天过期
	}, "7bdfc027-ef5f-67f3-af9f-311bcec930d5")
	t.Log(token)

	fromJwt, b := GetUserIdFromJwt(token, "7bdfc027-ef5f-67f3-af9f-311bcec930d5")
	t.Log(fromJwt, b)
}
