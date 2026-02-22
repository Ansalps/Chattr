TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MywiZW1haWwiOiJhbnNhbHBzOUBnbWFpbC5jb20iLCJyb2xlIjoidXNlciIsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJDaGF0dHIiLCJleHAiOjE3NzE2NTAwMDksImlhdCI6MTc3MTU2MzYwOSwianRpIjoiZjI5Yjc4N2ItMThkZi00ZDc4LWJjYmYtMWYwZjk0NmVkMzYzIn0.ZBqmP52YO8JPJdoC2OjMd377nPNxsVcf1OScIj5vRGw

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