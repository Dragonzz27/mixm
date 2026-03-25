from typing import List, Dict, Any

from app.models.user import User


class RouterService:
	"""Provide frontend route and menu configuration payloads."""

	def __init__(self):
		self._routes: List[Dict[str, Any]] = [
			{
				"path": "/login",
				"name": "Login",
				"component": "views/Login.vue",
				"meta": {"guest": True},
			},
			{"path": "/register", "redirect": "/login"},
			{
				"path": "/app",
				"component": "components/Layout.vue",
				"meta": {"requiresAuth": True, "isClient": True},
				"children": [
					{
						"path": "",
						"name": "ClientHome",
						"component": "views/client/Home.vue",
						"meta": {"title": "Home"},
					},
					{
						"path": "contents",
						"name": "ClientContents",
						"component": "views/client/Contents.vue",
						"meta": {"title": "内容列表"},
					},
					{
						"path": "levels",
						"name": "ClientLevels",
						"component": "views/client/Home.vue",
						"meta": {"title": "Levels"},
					},
					{
						"path": "levels/:id",
						"name": "ClientLevelDetail",
						"component": "views/client/Home.vue",
						"meta": {"title": "Level Detail"},
					},
					{
						"path": "my-instances",
						"name": "ClientMyInstances",
						"component": "views/client/MyInstances.vue",
						"meta": {"title": "我启动的实例"},
					},
					{
						"path": "instances/:id",
						"name": "ClientInstanceDetail",
						"component": "views/client/InstanceDetail.vue",
						"meta": {"title": "Instance Detail"},
					},
					{
						"path": "content/:id",
						"name": "ClientContent",
						"component": "views/client/ContentDetail.vue",
						"meta": {"title": "内容详情"},
					},
				],
			},
			{
				"path": "/admin",
				"component": "components/Layout.vue",
				"meta": {"requiresAuth": True, "isAdmin": True},
				"children": [
					{
						"path": "",
						"name": "Dashboard",
						"component": "views/admin/Dashboard.vue",
					},
					{
						"path": "users",
						"name": "Users",
						"component": "views/admin/Users.vue",
					},
					{
						"path": "categories",
						"name": "Categories",
						"component": "views/admin/Categories.vue",
					},
					{
						"path": "contents",
						"name": "Contents",
						"component": "views/admin/Contents.vue",
					},
					{
						"path": "templates",
						"name": "Templates",
						"component": "views/admin/Templates.vue",
					},
					{
						"path": "bindings",
						"name": "Bindings",
						"component": "views/admin/Bindings.vue",
					},
					{
						"path": "clusters",
						"name": "Clusters",
						"component": "views/admin/Clusters.vue",
					},
					{
						"path": "instances",
						"name": "Instances",
						"component": "views/admin/Instances.vue",
					},
					{
						"path": "images",
						"name": "AdminImageRepositories",
						"component": "views/admin/ImageRepositories.vue",
						"meta": {"title": "镜像仓库"},
					},
					{
						"path": "images/:id",
						"name": "AdminImageDetail",
						"component": "views/admin/ImageDetail.vue",
						"meta": {"title": "镜像详情"},
					},
				],
			},
			{"path": "/:pathMatch(.*)*", "redirect": "/app"},
		]

		self._menu: List[Dict[str, Any]] = [
			{"key": "ClientHome", "label": "首页", "path": "/app", "icon": "HomeOutline"},
			{
				"key": "ClientContents",
				"label": "内容浏览",
				"path": "/app/contents",
				"icon": "BookOutline",
			},
			{
				"key": "ClientMyInstances",
				"label": "我的实例",
				"path": "/app/my-instances",
				"icon": "CubeOutline",
			},
			{"type": "divider", "key": "d1", "label": ""},
			{"key": "Dashboard", "label": "仪表盘", "path": "/admin", "icon": "SpeedometerOutline"},
			{"key": "Users", "label": "用户管理", "path": "/admin/users", "icon": "PeopleOutline"},
			{
				"key": "Categories",
				"label": "分类管理",
				"path": "/admin/categories",
				"icon": "GridOutline",
			},
			{
				"key": "Contents",
				"label": "内容管理",
				"path": "/admin/contents",
				"icon": "DocumentTextOutline",
			},
			{
				"key": "Templates",
				"label": "模板管理",
				"path": "/admin/templates",
				"icon": "FileTrayFullOutline",
			},
			{
				"key": "Instances",
				"label": "实例管理",
				"path": "/admin/instances",
				"icon": "ServerOutline",
			},
			{
				"key": "AdminImageRepositories",
				"label": "镜像仓库",
				"path": "/admin/images",
				"icon": "ImageOutline",
			},
			{
				"key": "Clusters",
				"label": "集群管理",
				"path": "/admin/clusters",
				"icon": "LayersOutline",
			},
		]

	def get_routes(self, user: User) -> List[Dict[str, Any]]:
		# 按当前需求先返回全部路由，后续可在这里按角色裁剪。
		return self._routes

	def get_menu(self, user: User) -> List[Dict[str, Any]]:
		# 按当前需求先返回全部菜单，后续可在这里按角色裁剪。
		if user.is_superuser:
			return self._menu
		else:
			allow_menu_keys = ["ClientHome", "ClientContents", "ClientMyInstances"]
			return [item for item in self._menu if item.get("key") in allow_menu_keys]

