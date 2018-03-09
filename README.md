### Cache

Dead simple distributed cache with API over the REST. 

#### Features 

- Put value in cache
    - then all peers of the cache are sync with the state of the node 
- Get value from local map
- Delete the value in cache
    - then the value is deleted in all peers of the cache
- Value is the body of the PUT request
- Key is a string

#### Limitations 

- No consistency if one node breaks 
- No security
- In memory only - lack of persistency

#### Start

Build the executable cache app:
```
# ./dev.sh build
```

On each node
```
# ./cache -p 5555 -n=host1:5555,host2:5555,host3:5555
```

#### API

- PUT
```
curl -X PUT --data "VALUE" http://host1:5555/put/KEY
```

- GET
```
curl -X GET http://host1:5555/get/KEY
```

- DELETE
```
curl -X DELETE http://host1:5555/delete/KEY
```
