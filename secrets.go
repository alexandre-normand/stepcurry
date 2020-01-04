package stepcurry

type TokenLoader interface {
	LoadToken(teamID string) (token string, err error)
}

type TokenSaver interface {
	SaveToken(teamID string, token string) (err error)
}
