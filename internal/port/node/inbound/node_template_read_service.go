package inbound

type NodeTemplateReadService interface {
	List() ([]*NodeTemplateDTO, error)
}
