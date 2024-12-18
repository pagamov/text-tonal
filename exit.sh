# stop main app
docker ps -q --filter "ancestor=app" | xargs -r docker stop
# stop db
docker ps -q --filter "ancestor=db" | xargs -r docker stop
# stop redis
docker ps -q --filter "ancestor=redis:latest" | xargs -r docker stop