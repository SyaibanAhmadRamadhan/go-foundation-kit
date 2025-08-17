//go:generate go tool mockgen -destination=../../.mocking/hash_mock/hash_mock.go -package=hash_mock . Hasher
package hash

type Hasher interface {
	Hash(str string) (string, error)
	Verify(encodedHash string, str string) (bool, error)
}
