# Mancala Game

# Run Tests
```
go test ./...
```

## Run application
### Start redis
```
docker run -it --rm --name mancala-redis -p 6379:6379 redis:6.0-alpine
```

### Run Server
```
go run ./...
```

## How to Play the game

1) Run the server
2) Share the client with a friend
3) Each of you run the client (jq required)
    ```./client.sh http://{server_host}:8080```
4) Wait for your turn and select a pit from your board to make a move 

## Player Bot
```
for i in {1..1000}; do 
    n=`od -An -N1 -i /dev/random`; 
    n=$((n/25.5/2)); 
    echo $n | cut -d "." -f1; 
done | ./client.sh http://{server_host}:8080
```


## Docker build
```
VER=1.0
docker build --platform linux/amd64 -t mancala-server:$VER . && docker tag mancala-server:$VER registry.poiuytre.nl/mancala-server:$VER && docker push registry.poiuytre.nl/mancala-server:$VER
```

## Deploy to k8s
```
kubectl apply -f config.yaml
```
```
kubectl apply -f deployment.yaml
```
```
kubectl apply -f service.yaml
```  
