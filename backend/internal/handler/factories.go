package handler

import "mixm/backend/internal/service"

// NewResources 创建通用资源接口处理器。
func NewResources(s *service.Resources, spec service.ResourceSpec) *Resources {
	return &Resources{Service: s, spec: spec}
}

// NewImages 创建镜像仓库接口处理器并设置查询范围。
func NewImages(s *service.Images, mine bool) *Images {
	return &Images{Service: s, mine: mine}
}

// NewInstances 创建实例接口处理器并设置管理范围。
func NewInstances(s *service.Instances, all bool) *Instances {
	return &Instances{Service: s, all: all}
}

// NewBindings 创建内容绑定接口处理器并设置查询范围。
func NewBindings(s *service.Bindings, unbind bool) *Bindings {
	return &Bindings{Service: s, unbind: unbind}
}
