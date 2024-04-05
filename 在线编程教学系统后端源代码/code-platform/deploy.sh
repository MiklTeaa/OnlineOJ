git pull origin master

go build main.go

pgrep main | xargs kill -15

ulimit -c unlimited
export GOTRACEBACK=crash

./main &