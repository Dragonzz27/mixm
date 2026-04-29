from typing import List, Dict, Any, Optional, Set

from app.models.user import User


class RouterService:
    """Provide frontend route and menu configuration payloads."""

    def __init__(self):
        self._default_client_menu_keys = [
            "ClientHome",
            "ClientContents",
            "ClientMyInstances",
            "ClientAssignments",
        ]
        self._menu_key_route_names = {
            "ClientHome": ["ClientHome", "ClientLevels", "ClientLevelDetail"],
            "ClientContents": ["ClientContents", "ClientContent"],
            "ClientMyInstances": ["ClientMyInstances", "ClientInstanceDetail"],
            "ClientAssignments": ["ClientAssignments"],
            "Dashboard": ["Dashboard"],
            "Users": ["Users"],
            "Roles": ["Roles"],
            "Categories": ["Categories"],
            "Contents": ["Contents"],
            "Templates": ["Templates"],
            "Bindings": ["Bindings"],
            "Clusters": ["Clusters"],
            "Instances": ["Instances"],
            "AdminImageRepositories": ["AdminImageRepositories", "AdminImageDetail"],
            "Assignments": ["Assignments"],
            "Questions": ["Questions"],
        }

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
                    {
                        "path": "assignments",
                        "name": "ClientAssignments",
                        "component": "views/client/Assignments.vue",
                        "meta": {"title": "我的作业"},
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
                        "path": "roles",
                        "name": "Roles",
                        "component": "views/admin/Roles.vue",
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
                    {
                        "path": "assignments",
                        "name": "Assignments",
                        "component": "views/admin/Assignments.vue",
                        "meta": {"title": "作业管理"},
                    },
                    {
                        "path": "questions",
                        "name": "Questions",
                        "component": "views/admin/Questions.vue",
                        "meta": {"title": "题库管理"},
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
				"key": "ClientAssignments",
				"label": "我的作业",
				"path": "/app/assignments",
				"icon": "CreateOutline",
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
			{"key": "Roles", "label": "角色管理", "path": "/admin/roles", "icon": "ShieldCheckmarkOutline"},
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
			{
				"key": "Assignments",
				"label": "作业管理",
				"path": "/admin/assignments",
				"icon": "CreateOutline",
			},
			{
				"key": "Questions",
				"label": "题库管理",
				"path": "/admin/questions",
				"icon": "DocumentTextOutline",
			},
		]

    def _resolve_menu_keys(
        self, user: User, allow_menu_keys: Optional[List[str]]
    ) -> List[str]:
        if user.is_superuser:
            return [item.get("key") for item in self._menu if item.get("key")]
        if allow_menu_keys is None:
            return self._default_client_menu_keys
        return list(dict.fromkeys([key for key in allow_menu_keys if key]))

    def _resolve_route_names_by_menu_keys(self, allow_menu_keys: List[str]) -> Set[str]:
        allowed_route_names: Set[str] = set()
        for key in allow_menu_keys:
            for route_name in self._menu_key_route_names.get(key, [key]):
                allowed_route_names.add(route_name)
        return allowed_route_names

    def get_routes(
        self, user: User, allow_menu_keys: Optional[List[str]] = None
    ) -> List[Dict[str, Any]]:
        if user.is_superuser:
            return self._routes

        resolved_menu_keys = self._resolve_menu_keys(user, allow_menu_keys)
        allowed_route_names = self._resolve_route_names_by_menu_keys(resolved_menu_keys)
        filtered_routes: List[Dict[str, Any]] = []

        for route in self._routes:
            if route.get("path") in ["/login", "/register", "/:pathMatch(.*)*"]:
                filtered_routes.append(route)
                continue

            children = route.get("children")
            if not children:
                continue

            allowed_children = [
                child for child in children if child.get("name") in allowed_route_names
            ]
            if allowed_children:
                route_copy = dict(route)
                route_copy["children"] = allowed_children
                filtered_routes.append(route_copy)

        return filtered_routes

    def get_menu(
        self, user: User, allow_menu_keys: Optional[List[str]] = None
    ) -> List[Dict[str, Any]]:
        if user.is_superuser:
            return self._menu

        resolved_menu_keys = self._resolve_menu_keys(user, allow_menu_keys)
        return [item for item in self._menu if item.get("key") in resolved_menu_keys]

    def get_permission_options(self) -> List[Dict[str, str]]:
        permission_options: List[Dict[str, str]] = []
        for item in self._menu:
            key = item.get("key")
            if not key or item.get("type") == "divider":
                continue
            permission_options.append({"key": key, "label": item.get("label") or key})
        return permission_options
