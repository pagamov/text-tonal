echo "golang import get";

go get github.com/gin-gonic/gin;
go get github.com/jbrukh/bayesian;


echo "run build docker for main go app"

# RUN CGO_ENABLED=0 GOOS=linux go build -o /main main.go

# go build GOOS=linux 

# docker build -f app/main.Dockerfile -t main:v1 .

echo "setup postgresql for mac"

brew install postgresql

# pg_ctl -D /usr/local/opt/postgresql@14 start

/usr/local/opt/postgresql@14/bin/postgres -D /usr/local/var/postgresql@14

