package cryptography

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	//turn password into hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", ErrHashPassword
	}

	return string(hashedPassword), nil
}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
