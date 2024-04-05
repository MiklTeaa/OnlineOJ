dockerd &

# 让 Docker Daemon 启动完毕
for i in `seq 30`
do
    docker ps
    if [ $? -eq 0 ];
    then
        break
    fi
    sleep 1s
done


docker images -f reference=lgbgbl/monaco-python | grep -q monaco-python
if [ $? -ne 0 ];
then
    docker pull lgbgbl/monaco-python
fi

docker images -f reference=lgbgbl/monaco-java | grep -q monaco-java
if [ $? -ne 0 ];
then
    docker pull lgbgbl/monaco-java
fi

docker images -f reference=lgbgbl/monaco-cpp | grep -q monaco-cpp
if [ $? -ne 0 ];
then
    docker pull lgbgbl/monaco-cpp
fi

./server