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

docker images -f reference=lgbgbl/theia-python-auth | grep -q theia-python-auth
if [ $? -ne 0 ];
then
    docker pull lgbgbl/theia-python-auth
fi

docker images -f reference=lgbgbl/theia-java-auth | grep -q theia-java-auth
if [ $? -ne 0 ];
then
    docker pull lgbgbl/theia-java-auth
fi

docker images -f reference=lgbgbl/theia-cpp-auth | grep -q theia-cpp-auth
if [ $? -ne 0 ];
then
    docker pull lgbgbl/theia-cpp-auth
fi

./server