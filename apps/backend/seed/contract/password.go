package contract

import "golang.org/x/crypto/bcrypt"

const DemoPassword = "demo1234"

func DemoPasswordHash() string {
	hash, err := bcrypt.GenerateFromPassword([]byte(DemoPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}
