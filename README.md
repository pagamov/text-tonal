## app

[Контейнер с golang роутером](./app/Dockerfile)


Подключение к Redis

```golang
func initRedis() {
	ctx = context.Background()
	client = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",              
		DB:       0,               
	})
}
```

Добавление функционала

```golang
func (api *Router) Init() {
	// gin.SetMode(gin.ReleaseMode)
	api.router = gin.Default()
}

func (api *Router) AddMethod() {
	api.router.POST("/analyze", analyze)
	api.router.GET("/statistics/:begin/:end", statistics)
	api.router.GET("/ping", ping)
	api.router.GET("/logs", getLogs)
}

func (api *Router) Start(port string) {
	api.router.Run(fmt.Sprintf(":%s", port))
}
```


```golang

```


```golang

```


```golang

```


```golang

```


```golang

```

```golang

```

## db

## redis

## model

