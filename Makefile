TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MiwiZW1haWwiOiJhbnNhbHBzMTBAZ21haWwuY29tIiwicm9sZSI6InVzZXIiLCJ0eXBlIjoiYWNjZXNzIiwiaXNzIjoiQ2hhdHRyIiwiZXhwIjoxNzcxMzQzOTE3LCJpYXQiOjE3NzEyNTc1MTcsImp0aSI6IjY3NzRjMmE2LWNhN2UtNGNlNy04MWI5LWNiYjgxNTUxN2JkNCJ9.SvfahDXp34qEZDRbI9mX92Cr2hKt-vVRKOlWE7zbQYs

.PHONY: papiauth Run

papiauth:
	cd Api_Gateway && protoc --go_out=. --go-grpc_out=. ./pkg/proto/auth_subscription.proto
pauthauth:
	cd Auth_Subscription_Service && protoc --go_out=. --go-grpc_out=. ./pkg/pb/auth_subscription.proto
Run:
	cd Api_Gateway && go run ./cmd/main.go & \
	cd Auth_Subscription_Service && go run ./cmd/main.go & \
	cd Chat_Service && go run ./cmd/main.go & \
	cd Notification_Service && go run ./cmd/main.go & \
	cd Post_Relation_Service && go run ./cmd/main.go
Containers:
	cd Infrastructure/kafka && docker compose up -d & \
	docker start redis7 & \
	docker start mongodb
Kafkalog:
	docker exec kafka kafka-console-consumer \
	--bootstrap-server localhost:9092 \
	--include '.*-events' \
	--from-beginning
Kafkatopic:
	docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list
Chatws:
	wscat -c "ws://localhost:3000/user/ws" -H "Authorization: Bearer $(TOKEN)"
Notws:
	wscat -c "ws://localhost:3000/user/notification/ws" -H "Authorization: Bearer $(TOKEN)"