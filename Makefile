build:
	go build -trimpath -o mikrograf . 

docker_build:
	docker build -t sreng1neer/mikrograf:0.1.1 .

docker_push: docker_build
	docker push sreng1neer/mikrograf:0.1.1