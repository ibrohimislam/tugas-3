package server

import "fmt"
import "log"
import "strings"
import "crypto/rand"
import "crypto/sha1"
import "encoding/base64"
import "github.com/streadway/amqp"

type Server struct {
	Connection *amqp.Connection
	Users      map[string]User
	Groups     map[string]*Group
	Sessions   map[string]string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func NewServer(conn *amqp.Connection) Server {
	return Server{Users: make(map[string]User), Groups: make(map[string]*Group), Sessions: make(map[string]string), Connection: conn}
}

func (s *Server) Register(param []string) string {

	_, userExists := s.Users[param[0]]

	if userExists {

		return "{\"status\":\"failed\",\"message\":\"Registration failed, username already used.\"}"

	}

	h := sha1.New()
	h.Write([]byte(param[1]))
	s.Users[param[0]] = User{Username: param[0], Password: base64.URLEncoding.EncodeToString(h.Sum(nil))}

	return "{\"status\":\"success\",\"message\": \"Registration success.\"}"

}

func (s *Server) Login(param []string) string {

	user, userExists := s.Users[param[0]]

	if userExists {

		h := sha1.New()
		h.Write([]byte(param[1]))
		passwordSHA1 := base64.URLEncoding.EncodeToString(h.Sum(nil))

		if passwordSHA1 == user.Password {

			token := generateToken()
			s.Sessions[token] = param[0]
			return "{\"status\":\"success\",\"message\": \"Login success.\", \"token\": \"" + token + "\"}"

		}

	}

	return "{\"status\":\"failed\",\"message\":\"Login failed, username or password incorrect.\"}"

}

func (s *Server) AddFriend(token string, param []string) string {

	my_username := s.checkToken(token)

	if my_username == "" {
		return "{\"status\":\"failed\",\"message\": \"Token Invalid.\"}"
	}

	user, userExists := s.Users[param[0]]

	if userExists {

		s.sendToUser(user.Username, "system,"+my_username+" added You as friend.")

		return "{\"status\":\"success\",\"message\": \"Add Friend success.\"}"

	}

	return "{\"status\":\"failed\",\"message\": \"Add Friend failed, username not found.\"}"

}

func (s *Server) CreateGroup(token string, param []string) string {

	my_username := s.checkToken(token)

	if my_username == "" {
		return "{\"status\":\"failed\",\"message\": \"Token Invalid.\"}"
	}

	_, groupExists := s.Groups[param[0]]

	if groupExists {

		return "{\"status\":\"failed\",\"message\":\"Create Group failed, group already exists.\"}"

	}

	group := &Group{Name: param[0], AdminUsername: my_username, Members: []string{my_username}}

	s.Groups[param[0]] = group

	return "{\"status\":\"success\",\"message\": \"Create Group success.\"}"

}

func (s *Server) AddMember(token string, param []string) string {

	my_username := s.checkToken(token)

	if my_username == "" {

		return "{\"status\":\"failed\",\"message\": \"Token Invalid.\"}"

	}

	group, groupExists := s.Groups[param[0]]

	if !groupExists {

		return "{\"status\":\"failed\",\"message\":\"Add Member failed, group not exists.\"}"

	}

	if group.AdminUsername != my_username {

		return "{\"status\":\"failed\",\"message\":\"Add Member failed, only admin can add member.\"}"

	}

	user, userExists := s.Users[param[1]]

	if !userExists {

		return "{\"status\":\"failed\",\"message\":\"Add Member failed, username not found."

	}

	group = s.Groups[param[0]]

	group.Members = append(group.Members, param[1])

	s.sendToUser(user.Username, "system,"+my_username+" added You to group "+param[0]+".")

	return "{\"status\":\"success\",\"message\": \"Add Member success.\"}"

}

func (s *Server) RemoveMember(token string, param []string) string {

	my_username := s.checkToken(token)

	if my_username == "" {

		return "{\"status\":\"failed\",\"message\": \"Token Invalid.\"}"

	}

	group, groupExists := s.Groups[param[0]]

	if !groupExists {

		return "{\"status\":\"failed\",\"message\":\"Remove Member failed, group not exists.\"}"

	}

	if group.AdminUsername != my_username {

		return "{\"status\":\"failed\",\"message\":\"Remove Member failed, only admin can remove member.\"}"

	}

	user, userExists := s.Users[param[1]]

	if !userExists {

		return "{\"status\":\"failed\",\"message\":\"Remove Member failed, username not found."

	}

	members := group.Members

	b := members[:0]
	for _, username := range members {
		if username != param[1] {
			b = append(b, username)
		}
	}

	group.Members = b

	s.sendToUser(user.Username, "system,"+my_username+" removed You from group "+param[0]+".")

	return "{\"status\":\"success\",\"message\": \"Remove Member success.\"}"

}

func (s *Server) UserLeave(token string, param []string) string {

	my_username := s.checkToken(token)

	if my_username == "" {

		return "{\"status\":\"failed\",\"message\": \"Token Invalid.\"}"

	}

	group, groupExists := s.Groups[param[0]]

	if !groupExists {

		return "{\"status\":\"failed\",\"message\":\"Leave failed, group not exists.\"}"

	}

	members := group.Members

	b := members[:0]
	for _, username := range members {
		if username != my_username {
			b = append(b, username)
		}
	}

	group.Members = b

	s.sendToUser(my_username, "system, You leave from group "+param[0]+".")

	return "{\"status\":\"success\",\"message\": \"Leave Group success.\"}"

}

func (s *Server) UserSendToGroup(token string, param []string) string {

	my_username := s.checkToken(token)

	if my_username == "" {

		return "{\"status\":\"failed\",\"message\": \"Token Invalid.\"}"

	}

	group, groupExists := s.Groups[param[0]]

	if !groupExists {

		return "{\"status\":\"failed\",\"message\":\"Send Message failed, group not exists.\"}"

	}

	s.sendToGroup(group.Name, my_username+","+strings.Join(param[1:], " "))

	return "{\"status\":\"success\",\"message\": \"Send Message success.\"}"

}

func (s *Server) UserSendToUser(token string, param []string) string {

	my_username := s.checkToken(token)

	if my_username == "" {

		return "{\"status\":\"failed\",\"message\": \"Token Invalid.\"}"

	}

	user, userExists := s.Users[param[0]]

	if !userExists {

		return "{\"status\":\"failed\",\"message\":\"Send Message failed, username not found."

	}

	s.sendToUser(user.Username, my_username+","+strings.Join(param[1:], " "))

	return "{\"status\":\"success\",\"message\": \"Send Message success.\"}"

}

func (s *Server) sendToUser(username string, message string) {

	ch, err := s.Connection.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		username, // name
		true,     // durable
		false,    // delete when usused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("A," + message),
		},
	)
	failOnError(err, "Failed to publish a message")

}

func (s *Server) sendToGroup(groupname string, message string) {

	group := s.Groups[groupname]

	for _, username := range group.Members {

		ch, err := s.Connection.Channel()
		failOnError(err, "Failed to open a channel")
		defer ch.Close()

		q, err := ch.QueueDeclare(
			username, // name
			true,     // durable
			false,    // delete when usused
			false,    // exclusive
			false,    // no-wait
			nil,      // arguments
		)
		failOnError(err, "Failed to declare a queue")

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("B," + groupname + "," + message),
			},
		)
		failOnError(err, "Failed to publish a message")

	}

}

func generateToken() (uuid string) {

	b := make([]byte, 24)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
		return
	}

	uuid = fmt.Sprintf("%x", b)

	return
}

func (s *Server) checkToken(token string) string {
	username, tokenExists := s.Sessions[token]

	if tokenExists {

		return username

	}

	return ""
}
