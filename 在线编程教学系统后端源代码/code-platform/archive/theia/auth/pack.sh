node /home/theia/src-gen/backend/main.js /home/project "$@" &

THEIAPID=$!
sleep 3s
if kill -0 $THEIAPID > /dev/null 2> /dev/null; then
  cert="$CERTFILE" key="$KEYFILE" secure=$secure /usr/local/bin/gen-http-proxy localhost:3000
  kill $THEIAPID
else
  echo "could not spawn theia";
fi

