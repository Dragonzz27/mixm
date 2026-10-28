package handler

import "mixm/backend/internal/service"

func NewResources(s *service.Resources, spec service.ResourceSpec) *Resources {
	return &Resources{Service: s, spec: spec}
}
func NewImages(s *service.Images, mine bool) *Images         { return &Images{Service: s, mine: mine} }
func NewInstances(s *service.Instances, all bool) *Instances { return &Instances{Service: s, all: all} }
func NewBindings(s *service.Bindings, unbind bool) *Bindings {
	return &Bindings{Service: s, unbind: unbind}
}
