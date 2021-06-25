build:
	protoc -I="./../" --go_out="./" "./../proto/mealswipe/mealswipe.proto"
run:
	protoc -I="./../" --go_out="./" "./../proto/mealswipe/mealswipe.proto"
	docker build -t mealswipe .
	docker run -p 8080:8080 --name=mealswipe mealswipe 
stop:
	docker kill mealswipe
	docker rm mealswipe
test:
	cd src/mealswipe && go test -v ./...
tester:
	cd src/mealswipe && go run tester.go