package hash

import "golang.org/x/crypto/bcrypt"

// Password hashes a plaintext password with bcrypt (cost 12).
func Password(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ComparePassword returns true if plain matches the given bcrypt hash.
func ComparePassword(hashValue, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashValue), []byte(plain)) == nil
}
