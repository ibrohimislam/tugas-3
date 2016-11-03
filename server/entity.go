package server

type User struct {
	Username string
	Password string
}

type Group struct {
	Name          string
	AdminUsername string
	Members       []string
}
