
### server ###

```
cd main/itunneld
go build

./itunneld -httpAddr=:80 -tcpAddr=:222 -tunnelAddr=:4443
```

### client ###

```
cat << EOF >> ./config.yaml
server_addr: 127.0.0.1:4443
EOF

go build
./itunnel -config=./config.yaml -proto=tcp
```