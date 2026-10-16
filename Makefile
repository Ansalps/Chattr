TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6NSwiZW1haWwiOiJzb2JpamkzMTkzQG55c3ByaW5nLmNvbSIsInJvbGUiOiJ1c2VyIiwidHlwZSI6ImFjY2VzcyIsImlzcyI6IkNoYXR0ciIsImV4cCI6MTc3NjQyMDUxOSwiaWF0IjoxNzc2MzM0MTE5LCJqdGkiOiJmY2FmMzA5Mi03Y2ZkLTQ0MDMtOGIwNi1iMzg2ZjMzMTNmNWEifQ.kymd_SdSf3p8TnXPHa9wmJJqA5FKUv2tPhwcTgLtUOs

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
chatws:
	wscat -c "ws://localhost:3000/user/ws" -H "Authorization: Bearer $(TOKEN)"
notws:
	wscat -c "ws://localhost:3000/user/notification/ws" -H "Authorization: Bearer $(TOKEN)"
gkechatws:
	wscat -c "ws://chattr.shop/user/ws" -H "Authorization: Bearer $(TOKEN)"
gkenotws:
	wscat -c "ws://chattr.shop/user/notification/ws" -H "Authorization: Bearer $(TOKEN)"
dap:
	docker build -t ansalps/chattr-api_gateway ./Api_Gateway
klap:
	kind load docker-image chattr-api_gateway:latest --name single-node
krdap:
	kubectl rollout restart deployment api-gateway
das:
	docker build --no-cache -t chattr-auth_svc:latest ./Auth_Subscription_Service
klas:
	kind load docker-image chattr-auth_svc:latest --name single-node
krdas:
	kubectl rollout restart deployment auth-service
dps:
	docker build --no-cache -t chattr-post_svc:latest ./Post_Relation_Service
klps:
	kind load docker-image chattr-post_svc:latest --name single-node
krdps:
	kubectl rollout restart deployment post-relation-service
dcs:
	docker build -t chattr-chat_svc ./Chat_Service
klcs:
	kind load docker-image chattr-chat_svc:latest --name single-node
krdcs:
	kubectl rollout restart deployment chat-service
dns:
	docker build -t chattr-not_svc:latest ./Notification_Service
klns:
	kind load docker-image chattr-not_svc:latest --name single-node
krdns:
	kubectl rollout restart deployment notification-service

