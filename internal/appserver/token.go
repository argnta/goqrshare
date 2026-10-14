package appserver

import (
	"crypto/rand"
	"golang-share-files/internal/apperror"
	"math/big"
	"net"
	"net/url"
	"strconv"
)

func GenerateToken() (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	const n = 12
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", &apperror.Error{
				Kind: apperror.KindFailGenerateToken,
				Msg:  "Fail to Generate New Token",
				Err:  err,
			}
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func BuildURL(ip string, port int, token string) string {
	host := net.JoinHostPort(ip, strconv.Itoa(port))

	u := &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/" + token + "/",
	}

	return u.String()
}
