TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MywiZW1haWwiOiJhbnNhbHBzOUBnbWFpbC5jb20iLCJyb2xlIjoidXNlciIsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJDaGF0dHIiLCJleHAiOjE3NzQxOTk1NDYsImlhdCI6MTc3NDExMzE0NiwianRpIjoiMGVkYTk0ZmYtM2VkMi00ZDY3LTgxZGUtNWFiMGZlMzBiMzNkIn0.M9H4ZRr72S1M6kGqWL5fBV6WG0udomOjZgVnfG4B9bk

.PHONY: papiauth papipost Run

papiauth:
	cd Api_Gateway && protoc --go_out=. --go-grpc_out=. ./pkg/proto/auth_subscription.proto
papipost:
	cd Api_Gateway && protoc --go_out=. --go-grpc_out=. ./pkg/proto/post_relation.proto
pauthauth:
	cd Auth_Subscription_Service && protoc --go_out=. --go-grpc_out=. ./pkg/pb/auth_subscription.proto
ppostpost:
	cd Post_Relation_Service && protoc --go_out=. --go-grpc_out=. ./pkg/pb/post_relation.proto
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