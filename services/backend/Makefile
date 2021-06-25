build:
	protoc -I="./../" --go_out="./websocket-server/protobuf/" "./../protobuf/mealswipe/mealswipe.proto"
run:
	protoc -I="./../" --go_out="./websocket-server/protobuf/" "./../protobuf/mealswipe/mealswipe.proto"
	cd ./websocket-server/ && docker build -t mealswipe .
	docker run -p 8080:8080 --name=mealswipe mealswipe
stop:
	docker kill mealswipe
	docker rm mealswipe
test:
	cd ./websocket-server/ && go test -v ./...
tester:
	cd ./websocket-server/tools && go run tester.go