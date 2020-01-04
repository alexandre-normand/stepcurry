package stepcurry

// Connecter is implemented by any value that has a connect method
type Connecter interface {
	Connect() (err error)
}
