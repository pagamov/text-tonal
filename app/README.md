main file on port 8080

	api.router.POST("/analyze", analyze)
	api.router.GET("/statistics", statistics)
	api.router.GET("/ping", ping)