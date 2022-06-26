docker build . -t kind-app
kind create cluster --config=config/cluster.yml
kind load docker-image kind-app:latest
kubectl apply -f config/mysql-secret.yml
kubectl apply -f config/mysql.yml
kubectl apply -f config/app.yml
