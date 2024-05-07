#go-key-value-store

--Test Put HTTP
curl -X PUT -d'value to store' -v http://localhost:8080/api/v1/keyTest

--Test Get HTTP NoSuchKeyError
curl -v http://localhost:8080/api/v1/ketTest

--Test Get HtTP 
curl -v http://localhost:8080/api/v1/keyTest

--Test Delete HTTP
curl -X DELETE localhost:8080/api/v1/keyTest
