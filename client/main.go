package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func rpc(s string) (res string) {

	conn, err := amqp.Dial("amqp://guest:guest@ibrohim.me:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	corrId := randomString(32)

	err = ch.Publish(
		"",          // exchange
		"rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          []byte(s),
		})
	failOnError(err, "Failed to publish a message")

	for d := range msgs {
		if corrId == d.CorrelationId {
			res = string(d.Body)
			failOnError(err, "Failed to convert body to integer")
			break
		}
	}

	return
}

type response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

func main() {
	//rand.Seed(time.Now().UTC().UnixNano())

	args := os.Args

	if len(args) > 1 {

		s := strings.Join(args[1:], " ")
		res := rpc(s)
		fmt.Printf("%s\n", res)

		return

	}

	reader := bufio.NewReader(os.Stdin)

	var username string
	var token string

	username = ""

	var state int

	state = 0

	for {

		if state < 2 {
			fmt.Print("> ")
		} else {
			fmt.Print("[" + username + "] > ")
		}
		command, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		command = strings.TrimSpace(command)
		cmd := strings.Split(command, " ")

		var s string

		switch {
		case cmd[0] == "register":
			s = "{\"command\":\"register\",\"data\": \"" + cmd[1] + "," + cmd[2] + "\"}"
			break
		case cmd[0] == "login":
			s = "{\"command\":\"login\",\"data\": \"" + cmd[1] + "," + cmd[2] + "\"}"
			state = 1
			break
		case state == 2 && cmd[0] == "creategroup":
			s = "{\"command\":\"creategroup\",\"data\": \"" + cmd[1] + "\", \"token\":\"" + token + "\"}"
			break
		case state == 2 && cmd[0] == "addmember":
			s = "{\"command\":\"addmember\",\"data\": \"" + cmd[1] + "," + cmd[2] + "\", \"token\":\"" + token + "\"}"
			break
		case state == 2 && cmd[0] == "removemember":
			s = "{\"command\":\"removemember\",\"data\": \"" + cmd[1] + "," + cmd[2] + "\", \"token\":\"" + token + "\"}"
			break
		case state == 2 && cmd[0] == "sendtogroup":
			s = "{\"command\":\"sendtogroup\",\"data\": \"" + cmd[1] + "," + strings.Join(cmd[1:], " ") + "\", \"token\":\"" + token + "\"}"
			break
		case state == 2 && cmd[0] == "sendtouser":
			s = "{\"command\":\"sendtouser\",\"data\": \"" + cmd[1] + "," + strings.Join(cmd[1:], " ") + "\", \"token\":\"" + token + "\"}"
			break
		case state == 2 && cmd[0] == "leave":
			s = "{\"command\":\"leave\",\"data\": \"" + cmd[1] + "\", \"token\":\"" + token + "\"}"
			break

		case cmd[0] == "help":
			fallthrough

		default:
			fmt.Println("USAGE:")
			fmt.Println("-- before login --")
			fmt.Println("register [username] [password]")
			fmt.Println("login [username] [password]")
			fmt.Println("-- after login --")
			fmt.Println("creategroup [groupname]")
			fmt.Println("addmember [groupname] [username]")
			fmt.Println("removemember [groupname] [username]")
			fmt.Println("leave [groupname]")
			fmt.Println("sendtogroup [groupname] [message]")
			fmt.Println("sendtouser [username] [message]")

			continue
			break
		}

		res_json := rpc(s)

		if state == 1 {

			res := response{}
			json.Unmarshal([]byte(res_json), &res)

			if res.Status == "success" {
				state = 2
				username = cmd[1]
				token = res.Token

				go listenOnQueue(username)
			}

		}

		fmt.Printf("%s\n", res_json)
	}

}

func listenOnQueue(username string) {

	conn, err := amqp.Dial("amqp://guest:guest@ibrohim.me:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
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

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	for d := range msgs {
		splitted_msg := strings.Split(string(d.Body), ",")

		if splitted_msg[0] == "A" {

			message := strings.Join(splitted_msg[2:], ",")
			log.Printf("[direct] %s: %s", splitted_msg[1], message)

		} else {

			message := strings.Join(splitted_msg[3:], ",")
			log.Printf("[%s] %s: %s", splitted_msg[1], splitted_msg[2], message)

		}
	}

}
