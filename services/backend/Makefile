proto:
	protoc -I="./../" --go_out="./websocket-server/protobuf/" "./../protobuf/mealswipe/websocket.proto"
	protoc -I="./../" --go_out="./websocket-server/protobuf/" "./../protobuf/mealswipe/backend.proto"
run:
	protoc -I="./../" --go_out="./websocket-server/protobuf/" "./../protobuf/mealswipe/websocket.proto"
	cd ./websocket-server/ && docker build -f Dockerfile.wss -t mealswipe .
	docker run -p 8080:8080 --name=mealswipe mealswipe
deploy:
	protoc -I="./../" --go_out="./websocket-server/protobuf/" "./../protobuf/mealswipe/websocket.proto"
	cd ./websocket-server/ && docker build -f Dockerfile.wss -t mealswipe .
	docker tag mealswipe 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe
	docker push 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe
	kubectl rollout restart deployment mealswipe-ws
deploy-stats:
	protoc -I="./../" --go_out="./websocket-server/protobuf/" "./../protobuf/mealswipe/websocket.proto"
	cd ./websocket-server/ && docker build -f Dockerfile.stats -t mealswipe_stats .
	docker tag mealswipe_stats 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe-stats
	docker push 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe-stats
	kubectl rollout restart deployment mealswipe-stats
deploy-admin:
	protoc -I="./../" --go_out="./websocket-server/protobuf/" "./../protobuf/mealswipe/websocket.proto"
	cd ./websocket-server/ && docker build -f Dockerfile.admin -t mealswipe_admin .
	docker tag mealswipe_admin 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe-admin
	docker push 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe-admin
	kubectl rollout restart deployment mealswipe-admin
stop:
	docker kill mealswipe
	docker rm mealswipe
stop-nd:
	docker kill mealswipe
delete:
	docker rm mealswipe
test:
	cd ./websocket-server/ && go test -v ./...
tester:
	cd ./websocket-server/tools && go run tester.go
ecr-register:
	(Get-ECRLoginCommand).Password | docker login --username AWS --password-stdin 850351896280.dkr.ecr.us-east-1.amazonaws.com
mac-login:
	aws ecr get-login-password | docker login --username AWS --password-stdin 850351896280.dkr.ecr.us-east-1.amazonaws.com
publish:
	docker tag mealswipe 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe
	docker push 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe
kube-apply:
	cd ./websocket-server/build && kubectl apply -f .\mealswipe_websocket.yml
kube-apply-storage:
	cd ./websocket-server/build && kubectl apply -f .\redis_storageclass.yml
kubeapps:
	kubectl port-forward -n kubeapps svc/kubeapps 8080:80