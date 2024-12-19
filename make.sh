# docker main app
# -----------------------------
cd app/
docker build -t app .
docker run -d -p 8080:8080 app
cd ..
docker ps -q --filter "ancestor=app" | xargs -r echo "docker is up [app]"
# -----------------------------

# docker posgresql
# -----------------------------
cd db/
docker build -t db .
docker run -d -p 5432:5432 db
cd ..
docker ps -q --filter "ancestor=db" | xargs -r echo "docker is up [db]"
# -----------------------------

# now setup db and transfer data to posgresql
# -----------------------------
echo "[Copy backup data]"
cp ./backup/main.db.zip_part_* ./db/
rm db/main.db
echo "[Forming complete file]"
python3.12 ./db/main.py
echo "[Copy\`ing database]"
cd db/
go get github.com/mattn/go-sqlite3
go get github.com/lib/pq
go mod tidy
go run main.go
pgloader sqlite://main.db pgsql://pagamov:multipass@localhost/database
rm main.db
cd ..
# -----------------------------

# docker redis
# -----------------------------
docker pull redis:latest
docker run -d -p 6379:6379 redis:latest
docker ps -q --filter "ancestor=redis:latest" | xargs -r echo "docker is up [redis]"
# -----------------------------


# docker model
# -----------------------------
cd model/
docker build -t model .
docker run -d -p 8081:8081 model

# python3 -m venv venv
# source venv/bin/activate
# deactivate


cd ..
# -----------------------------
docker ps