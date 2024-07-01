package authentication

type Authenticator interface {
	Generate(string) (string, error)
	Verify(string) error
}
