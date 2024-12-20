1. start postgresql

This formula has created a default database cluster with:
  initdb --locale=C -E UTF-8 /usr/local/var/postgresql@14

To start postgresql@14 now and restart at login:
  brew services start postgresql@14
Or, if you don't want/need a background service you can just run:
  /usr/local/opt/postgresql@14/bin/postgres -D /usr/local/var/postgresql@14

2. start docker

To start colima now and restart at login:
  brew services start colima
Or, if you don't want/need a background service you can just run:
  /usr/local/opt/colima/bin/colima start -f



## start

colima start

docker build -t main:v1 .

<!-- docker run -it --rm main:v1 ls -l /build -->

<!-- docker images | grep main -->

<!-- docker run -it --rm main:v1 -->

docker run -it --rm -p 8080:8080 main:v1


<!-- Удаление зависших контейнеров -->
docker rmi $(sudo docker images -f “dangling=true” -q)

## stop

<!-- Удаление контейнера -->
docker image rm main:v1

colima stop
