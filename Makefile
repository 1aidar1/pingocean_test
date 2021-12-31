.PHONY:

example:
	go run app.go -workers 10 -time 4000 -needle "<div" https://google.com https://api.nutristeppe.com https://pingocean.com/ru https://vk.com https://hh.kz
build:
	go build app.go
test:
	go test
docker_build:
	docker build -t pingocean_cli .

	golangci-lint run

# make docker_run input="-needle link -workers 10 https://hh.kz"
input?=
docker_run:
	docker run --rm  pingocean_cli $(input)
docker_run_example:
	docker run --rm  pingocean_cli -needle link https://google.com https://hh.kz