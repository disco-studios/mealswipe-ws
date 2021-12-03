
run:
	docker build -f Dockerfile.wss -t mealswipe .
	docker run -p 8080:8080 --name=mealswipe mealswipe
deploy:
	docker build -f Dockerfile.wss -t mealswipe .
	docker tag mealswipe 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe
	docker push 850351896280.dkr.ecr.us-east-1.amazonaws.com/disco-mealswipe
	kubectl rollout restart deployment ms-ws
stop:
	docker kill mealswipe
	docker rm mealswipe
stop-nd:
	docker kill mealswipe
delete:
	docker rm mealswipe
tester:
	cd tools && go run tester.go
ecr-register:
	(Get-ECRLoginCommand).Password | docker login --username AWS --password-stdin 850351896280.dkr.ecr.us-east-1.amazonaws.com
mac-login:
	aws ecr get-login-password | docker login --username AWS --password-stdin 850351896280.dkr.ecr.us-east-1.amazonaws.com