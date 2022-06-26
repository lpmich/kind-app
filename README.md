## Kind App
A simple golang web application deployed through [Kind Kubernetes](https://kind.sigs.k8s.io/)

### Prerequisites
- WSL2 or linux shell
- Docker Desktop
- Go
- Kind

### Installation
```bash
git pull https://github.com/lpmich/kind-app.git
cd kind-app/
```

### Configuration
```bash
vi config/mysql-secret.yml
```
Insert the following code into the config/mysql-secret.yml, enter a desired password for the mysql database into the password field and then save the file with :wq
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysql-secret
type: kubernetes.io/basic-auth
stringData:
  password:
```

### Run the Application
```bash
./build.sh
```
After the script finishes the application should be running at http://localhost
