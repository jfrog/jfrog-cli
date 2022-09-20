docker build -t artifactory .
docker network prune -f
docker network create "test-network"
docker run -p 8082:8082  -p 8081:8081 --name artifactory --network "test-network" --env RTLIC="$RTLIC" --env GOPROXY="direct" artifactory
