.PHONY: build run test clean docker-build docker-push k8s-deploy k8s-clean

# Go commands
build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v

clean:
	rm -rf bin/

# Docker commands
docker-build:
	docker build -t go-ai-service:latest .

docker-run:
	docker run -p 8080:8080 --env-file .env go-ai-service:latest

# Minikube commands
minikube-start:
	minikube start --cpus=4 --memory=8192 --disk-size=20gb --driver=docker

minikube-load:
	minikube image load go-ai-service:latest

# Kubernetes commands
k8s-deploy:
	kubectl apply -f kubernetes/

k8s-status:
	kubectl get all -n default

k8s-logs:
	kubectl logs -l app=go-ai-service -f

k8s-clean:
	kubectl delete -f kubernetes/

# Monitoring
install-monitoring:
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo update
	helm install prometheus prometheus-community/prometheus --namespace monitoring --create-namespace
	helm install grafana grafana/grafana --namespace monitoring --set service.type=NodePort --set adminPassword=admin

# Load testing
load-test:
	locust -f locustfile.py --host=http://localhost:8080 --users=1000 --spawn-rate=100 --run-time=5m