package auth

type Hasher interface {
	Hash(text string) (string, error)
	Verify(hashedText, text string) error
}
