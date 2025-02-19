# OSRM Traffic Updater
The **OSRM Traffic Updater** is designed for pull traffic data from **Traffic Proxy(Telenav)** then dump to OSRM required `traffic.csv`. Refer to [OSRM with Telenav Traffic Design](../docs/design/osrm-with-telenav-traffic.md) and [OSRM Traffic](https://github.com/Project-OSRM/osrm-backend/wiki/Traffic) for more details.        
We have implemented both `Python` and `Go` version. Both of them have same function(pull data then dump to csv), but the `Go` implementation is about **23 times** faster than `Python` implementation. So strongly recommended to use `Go` implementation as preference.        
- E.g. `6727490` lines traffic of NA region     
    - `Go` Implementation: about `9 seconds`
    - `Python` Implementation: about `210 seconds`    

## RPC Protocol
See [proxy.thrift](proxy.thrift) for details.    

## Python Implementation
The `python` based implementation has been deprecated due to bad performance. See [Deprecated Python Implementation Codes](https://github.com/Telenav/osrm-backend/blob/b4eb73f2d307fd4dbd8b8610bbc2a68c3b6ab1ae/traffic_updater/python/osrm_traffic_updater.py#L57) if you'd like to see code details.     

## Go Implementation
### Requirements
- `go version go1.12.5 linux/amd64`
- `thrift 0.12.0`
    - clone `thrift` from `github.com/apache/thrift`, then checkout branch `0.12.0`
- change `thrift` imports in generated codes `gen-go/proxy` 
    - `git.apache.org/thrift.git/lib/go/thrift` -> `github.com/apache/thrift/lib/go/thrift`


### Usage
```bash
$ cd $GOPATH
$ go install github.com/Telenav/osrm-backend/traffic_updater/go/osrm_traffic_updater
$ ./bin/osrm_traffic_updater -h
Usage of ./bin/osrm_traffic_updater:
  -c string
        traffic proxy ip address (default "127.0.0.1")
  -d    use high precision speeds, i.e. decimal (default true)
  -f string
        OSRM traffic csv file (default "traffic.csv")
  -p int
        traffic proxy listening port (default 6666)
```

