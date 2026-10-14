package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"mixm-edu/backend/internal/apperror"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/handler"
	"mixm-edu/backend/internal/infra/cluster"
	"mixm-edu/backend/internal/middleware"
	"mixm-edu/backend/internal/repository"
	"mixm-edu/backend/internal/service"
)

type Router struct {
	Engine *gin.Engine
	config config.Config
	db     *repository.DB
}

func New(c config.Config, db *repository.DB, cache *redis.Client, k *cluster.Client) *Router {
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
	protected.GET("/users/me/submissions", assignmentsHandler.Mine)
	protected.POST("/assignments/:id/submissions", assignmentsHandler.Submit)
	protected.GET("/assignments/:id/submissions", assignmentsHandler.Submissions)
	protected.PATCH("/submissions/:id", assignmentsHandler.Review)
	protected.DELETE("/submissions/:id", assignmentsHandler.DeleteSubmission)
	protected.GET("/assignments/:id/submissions/export", assignmentsHandler.Export)
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
	protected.POST("/bindings", bindingsHandler.Create)
	protected.GET("/bindings", bindingsHandler.List)
	protected.GET("/bindings/:id", bindingsHandler.Detail)
	protected.DELETE("/bindings/:id", bindingsHandler.Delete)
	protected.DELETE("/contents/:id/category-bindings", bindingsHandler.DeleteForContent)
	protected.POST("/categories/:id/content-bindings", bindingsHandler.Batch)
	protected.GET("/binding-counts", bindingsHandler.Counts)
	protected.GET("/categories/:id/contents", bindingsHandler.CategoryContents)
	protected.GET("/categories/:id/content-tree", bindingsHandler.ContentTree)
	protected.GET("/clusters/pods", clusterHandler.GetPods)
	protected.GET("/clusters/nodes", clusterHandler.GetNodes)
	protected.GET("/clusters/deployments", clusterHandler.GetDeployments)
	protected.DELETE("/clusters/namespaces/:namespace/deployments/:name", clusterHandler.DeleteDeployments)
	protected.DELETE("/clusters/namespaces/:namespace/pods/:name", clusterHandler.DeletePods)
	protected.GET("/clusters/namespaces/:namespace/pods/:name/logs", clusterHandler.GetPodsLogs)
	protected.GET("/clusters/namespaces/:namespace/pods/:name/describe", clusterHandler.GetPodsDescribe)
	protected.GET("/clusters/persistent-volumes", clusterHandler.GetPv)
	protected.DELETE("/clusters/persistent-volumes/:name", clusterHandler.DeletePv)
	protected.GET("/clusters/persistent-volumesc", clusterHandler.GetPvc)
	protected.DELETE("/clusters/namespaces/:namespace/persistent-volume-claims/:name", clusterHandler.DeletePvc)
	protected.POST("/assignments/files", filesHandler.AssignmentUpload)
	public.GET("/assignments/files/:filename", filesHandler.AssignmentDownload)
	protected.POST("/contents/files", filesHandler.ContentUpload)
	public.GET("/contents/files/*key", filesHandler.ContentDownload)
	protected.GET("/images", imagesHandler.List)
	protected.POST("/images", imagesHandler.Create)
	protected.GET("/images/:id", imagesHandler.Detail)
	protected.GET("/images/:id/tags", imagesHandler.Tags)
	protected.PATCH("/images/:id", imagesHandler.Update)
	protected.DELETE("/images/:id", imagesHandler.Delete)
	protected.DELETE("/images/:id/tags/:tag", imagesHandler.DeleteTag)
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
	protected.POST("/practices/generations", learningHandler.PracticeGeneration)
	protected.POST("/simulation-attempts", learningHandler.Generate)
	protected.POST("/simulation-attempts/:id/submissions", learningHandler.Submit)
	protected.GET("/simulation-attempts/current", learningHandler.Current)
	protected.GET("/simulation-attempts/:id/review", learningHandler.Review)
	protected.GET("/simulation-traces", learningHandler.Traces)
	protected.GET("/knowledge-tracing/status", learningHandler.KTStatus)
	protected.POST("/questions/generations", learningHandler.QuestionGeneration)
	protected.POST("/knowledge-tree/analyses", learningHandler.KnowledgeAnalysis)
	protected.POST("/knowledge-tree", learningHandler.SaveKnowledge)
	protected.GET("/categories/tree", resourcesHandler.CategoryTree)
	protected.GET("/question-categories/tree", resourcesHandler.QuestionCategoryTree)
	protected.POST("/questions/batches", resourcesHandler.BatchQuestions)
	protected.PUT("/questions/:id/review", resourcesHandler.ReviewQuestion)
	protected.GET("/system/stats", systemHandler.Stats)
	protected.GET("/system/load", systemHandler.Load)
	for _, spec := range service.ResourcesList {
		h := handler.NewResources(resources, spec)
		group := protected.Group("/" + spec.Path)
		group.GET("", h.List)
		group.POST("", h.Create)
		group.GET("/:id", h.Detail)
		group.PATCH("/:id", h.Save)
		group.DELETE("/:id", h.Delete)
	}
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

	router := &Router{Engine: engine, config: c, db: db}
	engine.GET("/health", router.Health)
	engine.GET("/static/*filepath", router.Static)
	engine.HEAD("/static/*filepath", router.Static)
	engine.GET("/openapi.json", router.OpenAPI)
	engine.GET("/docs", router.Docs)
	return router
}
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
func (r *Router) Health(c *gin.Context) {
	if r.db == nil || r.db.Ping(c.Request.Context()) != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "数据库不可用"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
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
