package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"mixm/backend/internal/apperror"
	"mixm/backend/internal/config"
	"mixm/backend/internal/handler"
	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/middleware"
	"mixm/backend/internal/repository"
	"mixm/backend/internal/service"
)

type Router struct {
	Engine *gin.Engine
	config config.Config
	db     *repository.DB
}

// New 组装服务与接口处理器，注册公共和受保护的路由。
func New(c config.Config, db *repository.DB, cache *redis.Client, k *cluster.Client) *Router {
	// 初始化公共中间件及统一的路由错误响应。
	engine := gin.New()
	engine.RedirectTrailingSlash = false
	engine.RedirectFixedPath = false
	engine.HandleMethodNotAllowed = true
	_ = engine.SetTrustedProxies(nil)
	engine.Use(middleware.CORS, middleware.Logger, middleware.Recovery)
	engine.NoRoute(func(c *gin.Context) {
		handler.Failure(c, &apperror.Error{Status: http.StatusNotFound, Detail: "资源不存在"})
	})
	engine.NoMethod(func(c *gin.Context) {
		handler.Failure(c, &apperror.Error{Status: http.StatusMethodNotAllowed, Detail: "请求方法不允许"})
	})

	// 构造服务依赖，并区分公开接口和需要认证的接口。
	core := service.NewCore(c, db, cache, k)
	public := engine.Group("/api")
	protected := engine.Group("/api", middleware.Auth(c, db))
	assignments := &service.Assignments{Core: core}
	bindings := &service.Bindings{Core: core}
	cluster := &service.Cluster{Core: core}
	files := &service.Files{Core: core}
	images := &service.Images{Core: core}
	instances := &service.Instances{Core: core}
	learning := &service.Learning{Core: core}
	resources := &service.Resources{Core: core}
	system := &service.System{Core: core}
	users := &service.Users{Core: core}

	// 为各业务服务创建对应的 HTTP 处理器。
	assignmentsHandler := &handler.Assignments{Service: assignments}
	clusterHandler := &handler.Cluster{Service: cluster}
	filesHandler := &handler.Files{Service: files}
	learningHandler := &handler.Learning{Service: learning}
	systemHandler := &handler.System{Service: system}
	usersHandler := &handler.Users{Service: users}
	resourcesHandler := handler.NewResources(resources, service.ResourceSpec{})
	imagesHandler := handler.NewImages(images, false)
	instancesHandler := handler.NewInstances(instances, true)
	bindingsHandler := handler.NewBindings(bindings, false)

	// 作业提交、批改和导出接口。
	protected.GET("/users/me/submissions", assignmentsHandler.Mine)
	protected.POST("/assignments/:id/submissions", assignmentsHandler.Submit)
	protected.GET("/assignments/:id/submissions", assignmentsHandler.Submissions)
	protected.PATCH("/submissions/:id", assignmentsHandler.Review)
	protected.DELETE("/submissions/:id", assignmentsHandler.DeleteSubmission)
	protected.GET("/assignments/:id/submissions/export", assignmentsHandler.Export)

	// 用户账号、登录令牌和角色权限接口。
	protected.POST("/users", usersHandler.Create)
	public.POST("/auth/login", usersHandler.Login)
	public.POST("/auth/refresh", usersHandler.Refresh)
	protected.GET("/users", usersHandler.List)
	protected.GET("/users/:id", usersHandler.Detail)
	protected.PATCH("/users/:id", usersHandler.Update)
	protected.PUT("/users/:id/password", usersHandler.Password)
	protected.DELETE("/users/:id", usersHandler.Delete)
	protected.POST("/users/imports", usersHandler.Import)
	protected.GET("/roles/permissions", usersHandler.Permissions)

	// 分类内容绑定及查询接口。
	protected.POST("/bindings", bindingsHandler.Create)
	protected.GET("/bindings", bindingsHandler.List)
	protected.GET("/bindings/:id", bindingsHandler.Detail)
	protected.DELETE("/bindings/:id", bindingsHandler.Delete)
	protected.DELETE("/contents/:id/category-bindings", bindingsHandler.DeleteForContent)
	protected.POST("/categories/:id/content-bindings", bindingsHandler.Batch)
	protected.GET("/binding-counts", bindingsHandler.Counts)
	protected.GET("/categories/:id/contents", bindingsHandler.CategoryContents)
	protected.GET("/categories/:id/content-tree", bindingsHandler.ContentTree)

	// 集群资源、容器日志和持久卷管理接口。
	protected.GET("/clusters/pods", clusterHandler.GetPods)
	protected.GET("/clusters/nodes", clusterHandler.GetNodes)
	protected.GET("/clusters/deployments", clusterHandler.GetDeployments)
	protected.DELETE("/clusters/namespaces/:namespace/deployments/:name", clusterHandler.DeleteDeployments)
	protected.DELETE("/clusters/namespaces/:namespace/pods/:name", clusterHandler.DeletePods)
	protected.GET("/clusters/namespaces/:namespace/pods/:name/logs", clusterHandler.GetPodsLogs)
	protected.GET("/clusters/namespaces/:namespace/pods/:name/describe", clusterHandler.GetPodsDescribe)
	protected.GET("/clusters/persistent-volumes", clusterHandler.GetPv)
	protected.DELETE("/clusters/persistent-volumes/:name", clusterHandler.DeletePv)
	protected.GET("/clusters/persistent-volume-claims", clusterHandler.GetPvc)
	protected.DELETE("/clusters/namespaces/:namespace/persistent-volume-claims/:name", clusterHandler.DeletePvc)

	// 附件上传与公开下载接口。
	protected.POST("/assignments/files", filesHandler.AssignmentUpload)
	public.GET("/assignments/files/:filename", filesHandler.AssignmentDownload)
	protected.POST("/contents/files", filesHandler.ContentUpload)
	public.GET("/contents/files/*key", filesHandler.ContentDownload)

	// 镜像仓库与标签管理接口。
	protected.GET("/images", imagesHandler.List)
	protected.POST("/images", imagesHandler.Create)
	protected.POST("/images/sync", imagesHandler.Sync)
	protected.GET("/images/:id", imagesHandler.Detail)
	protected.GET("/images/:id/tags", imagesHandler.Tags)
	protected.PATCH("/images/:id", imagesHandler.Update)
	protected.DELETE("/images/:id", imagesHandler.Delete)
	protected.DELETE("/images/:id/tags/:tag", imagesHandler.DeleteTag)

	// 实例生命周期和持久化创建任务接口。
	protected.POST("/instances", instancesHandler.Create)
	protected.GET("/instances", instancesHandler.List)
	protected.GET("/instances/:id", instancesHandler.Detail)
	protected.PATCH("/instances/:id", instancesHandler.Update)
	protected.DELETE("/instances/:id", instancesHandler.Delete)
	protected.DELETE("/instances", instancesHandler.DeleteAll)
	protected.POST("/templates/:id/instances", instancesHandler.Batch)
	protected.GET("/instance-tasks/:id", instancesHandler.Task)
	protected.GET("/instance-tasks", instancesHandler.Tasks)
	protected.GET("/instance-tasks/summary", instancesHandler.Summary)

	// 练习、测评及大模型生成接口。
	protected.POST("/practices/generations", learningHandler.PracticeGeneration)
	protected.POST("/simulation-attempts", learningHandler.Generate)
	protected.POST("/simulation-attempts/:id/submissions", learningHandler.Submit)
	protected.GET("/simulation-attempts/current", learningHandler.Current)
	protected.GET("/simulation-attempts/:id/review", learningHandler.Review)
	protected.POST("/questions/generations", learningHandler.QuestionGeneration)
	protected.POST("/knowledge-tree/analyses", learningHandler.KnowledgeAnalysis)
	protected.POST("/knowledge-tree", learningHandler.SaveKnowledge)

	// 分类树、题目审核和宿主机状态接口。
	protected.GET("/categories/tree", resourcesHandler.CategoryTree)
	protected.GET("/question-categories/tree", resourcesHandler.QuestionCategoryTree)
	protected.POST("/questions/batches", resourcesHandler.BatchQuestions)
	protected.PUT("/questions/:id/review", resourcesHandler.ReviewQuestion)
	protected.GET("/system/stats", systemHandler.Stats)
	protected.GET("/system/load", systemHandler.Load)

	// 为通用资源注册一致的增删改查接口。
	for _, spec := range service.ResourcesList {
		h := handler.NewResources(resources, spec)
		group := protected.Group("/" + spec.Path)
		group.GET("", h.List)
		group.POST("", h.Create)
		group.GET("/:id", h.Detail)
		group.PATCH("/:id", h.Save)
		group.DELETE("/:id", h.Delete)
	}

	// 补充用户范围查询，并保留历史附件地址与 HEAD 请求。
	protected.GET("/users/me/images", handler.NewImages(images, true).List)
	protected.GET("/users/me/instances", handler.NewInstances(instances, false).Mine)
	protected.GET("/users/:id/instances", handler.NewInstances(instances, false).ForUser)
	protected.DELETE("/categories/:id/content-bindings", handler.NewBindings(bindings, true).Batch)
	public.GET("/assignments/file/:filename", filesHandler.AssignmentDownload)
	public.GET("/contents/file/*key", filesHandler.ContentDownload)
	public.HEAD("/assignments/files/:filename", filesHandler.AssignmentDownload)
	public.HEAD("/assignments/file/:filename", filesHandler.AssignmentDownload)
	public.HEAD("/contents/files/*key", filesHandler.ContentDownload)
	public.HEAD("/contents/file/*key", filesHandler.ContentDownload)

	// 注册健康检查、静态文件和 API 文档入口。
	router := &Router{
		Engine: engine,
		config: c,
		db:     db,
	}
	engine.GET("/health", router.Health)
	engine.GET("/static/*filepath", router.Static)
	engine.HEAD("/static/*filepath", router.Static)
	engine.GET("/openapi.json", router.OpenAPI)
	engine.GET("/docs", router.Docs)

	return router
}

// ServeHTTP 通过统一路由处理 HTTP 请求。
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) > 1 {
		req.URL.Path = strings.TrimSuffix(req.URL.Path, "/")
	}

	if req.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.WriteHeader(http.StatusNoContent)

		return
	}

	r.Engine.ServeHTTP(w, req)
}

// Health 返回应用健康检查结果。
func (r *Router) Health(c *gin.Context) {
	if r.db == nil || r.db.Ping(c.Request.Context()) != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "数据库不可用"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Static 安全解析并返回本地静态文件。
func (r *Router) Static(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("filepath"), "/")
	directory := "static"
	if strings.HasPrefix(path, "export/") {
		path = strings.TrimPrefix(path, "export/")
		directory = r.config.ExportDir
	}

	req := c.Request.Clone(c.Request.Context())
	req.URL.Path = "/" + path
	http.FileServer(http.Dir(directory)).ServeHTTP(c.Writer, req)
}
