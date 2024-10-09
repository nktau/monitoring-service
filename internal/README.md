# internal

В данной директории и её поддиректориях будет содержаться имплементация вашего сервиса



```
 curl -X POST http://localhost:8080/update/ -H "Content-Type:application/json" -d "{\"id\":\"test_name\",\"type\":\"gauge\",\"value\":234}"
```

```
curl -X POST http://localhost:8080/updates/ -H "Content-Type:application/json" -d "[{\"id\":\"test_name\",\"type\":\"gauge\",\"value\":234},{\"id\":\"test_name1\",\"type\":\"gauge\",\"value\":2332}]"
```

```
cd agent/ && go build . && cd -
```

```
 curl -X POST http://localhost:8080/value/ -H "Content-Type:application/json" -d "{\"id\":\"test_name\",\"type\":\"gauge\"}"

```

