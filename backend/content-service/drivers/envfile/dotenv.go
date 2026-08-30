package envfile

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadDotEnv(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{".env"}
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if err := godotenv.Load(p); err != nil {
			return err
		}
	}
	return nil
}
