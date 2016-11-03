go run main.go {\"command\":\"register\",\"data\": \"A,13513090\"}
go run main.go {\"command\":\"register\",\"data\": \"B,13513090\"}

token=$(go run main.go {\"command\":\"login\",\"data\": \"A,13513090\"} | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["token"]')

go run main.go {\"command\":\"creategroup\",\"data\": \"if4031\",\"token\":\"$token\"}
go run main.go {\"command\":\"addmember\",\"data\": \"if4031,B\",\"token\":\"$token\"}
go run main.go {\"command\":\"sendtogroup\",\"data\": \"if4031,lalalalala\",\"token\":\"$token\"}
