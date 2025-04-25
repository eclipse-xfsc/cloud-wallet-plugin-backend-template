include $(PWD)/.env

build:
	docker build -f deployment/docker/Dockerfile -t (IMAGE_REPO):$(IMAGE_TAG) --env-file=.env .

restart: build
	docker run -d -p 8080:8080 --env-file=.env --name plugin-backend (IMAGE_REPO):$(IMAGE_TAG)

update: build
	docker push $(IMAGE_REPO):$(IMAGE_TAG)
