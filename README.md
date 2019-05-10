Do:
- Define middlewares with name
- Define each route as tree:

```
{
    "middlewares": {
        "api": [ "ValidationMW", "LogMW"],
        "root": [ "LogMW" ]
    }

    "matchers" {

    }

    "routes": {
        "/api" : {
            "/test": {
                "Methods": [ "GET" ],
                "middlewares": { "$ref": "#/middlewares/api" }
            }
        }
    }
}
```

```
Node Root
    Node Middlewares
        Node api
            Node ValidationMW
            Node LogMW
        Node root
            Node LogMW

```

Rules: 
- Middleware/Matcher rules must only define functions / array of functions
- Middleware/Matcher rules may only be referenced in Routes
