//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2017] Last.Backend LLC
// All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains
// the property of Last.Backend LLC and its suppliers,
// if any.  The intellectual and technical concepts contained
// herein are proprietary to Last.Backend LLC
// and its suppliers and may be covered by Russian Federation and Foreign Patents,
// patents in process, and are protected by trade secret or copyright law.
// Dissemination of this information or reproduction of this material
// is strictly forbidden unless prior written permission is obtained
// from Last.Backend LLC.
//

package storage

import (
	"context"
	"github.com/lastbackend/lastbackend/pkg/apis/types"
	"golang.org/x/oauth2"
)

type IUtil interface {
	Key(ctx context.Context, pattern ...string) string
}

type IStorage interface {
	Activity() IActivity
	Build() IBuild
	Hook() IHook
	Image() IImage
	Project() INamespace
	Service() IService
	Vendor() IVendor
	Volume() IVolume
}

type IActivity interface {
	Insert(ctx context.Context, activity *types.Activity) (*types.Activity, error)
	ListProjectActivity(ctx context.Context, id string) (*types.ActivityList, error)
	ListServiceActivity(ctx context.Context, id string) (*types.ActivityList, error)
	RemoveByProject(ctx context.Context, id string) error
	RemoveByService(ctx context.Context, id string) error
}

type IBuild interface {
	GetByID(ctx context.Context, imageID, id string) (*types.Build, error)
	ListByImage(ctx context.Context, id string) (*types.BuildList, error)
	Insert(ctx context.Context, imageID string, source *types.BuildSource) (*types.Build, error)
}

type IHook interface {
	GetByToken(ctx context.Context, token string) (*types.Hook, error)
	ListByImage(ctx context.Context, id string) (*types.HookList, error)
	ListByService(ctx context.Context, id string) (*types.HookList, error)
	Insert(ctx context.Context, hook *types.Hook) (*types.Hook, error)
	Remove(ctx context.Context, id string) error
	RemoveByService(ctx context.Context, id string) error
}

type INamespace interface {
	GetByID(ctx context.Context, id string) (*types.Namespace, error)
	GetByName(ctx context.Context, name string) (*types.Namespace, error)
	List(ctx context.Context) (*types.NamespaceList, error)
	Insert(ctx context.Context, name, description string) (*types.Namespace, error)
	Update(ctx context.Context, project *types.Namespace) (*types.Namespace, error)
	Remove(ctx context.Context, id string) error
}

type IService interface {
	GetByID(ctx context.Context, namespace, id string) (*types.Service, error)
	GetByName(ctx context.Context, namespace string, name string) (*types.Service, error)
	GetByPodID(ctx context.Context, uuid string) (*types.Service, error)
	ListByProject(ctx context.Context, namespace string) (*types.ServiceList, error)
	Insert(ctx context.Context, namespace string, name, description string, config *types.ServiceConfig) (*types.Service, error)
	Update(ctx context.Context, namespace string, service *types.Service) (*types.Service, error)
	Remove(ctx context.Context, namespace string, service *types.Service) error
	RemoveByProject(ctx context.Context, namespace string) error
}

type IPod interface {
	GetByID(ctx context.Context, namespace, service, id string) (*types.PodNodeState, error)
	ListByService(ctx context.Context, namespace, service string) ([]*types.PodNodeState, error)
	Insert(ctx context.Context, namespace, service string, pod *types.PodNodeState) error
	Update(ctx context.Context, namespace, service string, pod *types.PodNodeState) error
	Remove(ctx context.Context, namespace, service string, pod *types.PodNodeState) error
}

type IImage interface {
	Get(ctx context.Context, name string) (*types.Image, error)
	Insert(ctx context.Context, name string, source *types.ImageSource) (*types.Image, error)
	Update(ctx context.Context, image *types.Image) (*types.Image, error)
}

type IVendor interface {
	Insert(ctx context.Context, owner, name, host, serviceID string, token *oauth2.Token) error
	Get(ctx context.Context, name string) (*types.Vendor, error)
	List(ctx context.Context) (map[string]*types.Vendor, error)
	Update(ctx context.Context, vendor *types.Vendor) error
	Remove(ctx context.Context, vendorName string) error
}

type IVolume interface {
	GetByToken(ctx context.Context, token string) (*types.Volume, error)
	ListByProject(ctx context.Context, project string) (*types.VolumeList, error)
	Insert(ctx context.Context, volume *types.Volume) (*types.Volume, error)
	Remove(ctx context.Context, id string) error
}

type INode interface {
	List(ctx context.Context) (*types.NodeList, error)

	Get(ctx context.Context, hostname string) (*types.Node, error)
	Insert(ctx context.Context, meta *types.NodeMeta, state *types.NodeState) (*types.Node, error)

	UpdateMeta(ctx context.Context, meta *types.NodeMeta) error
	UpdateState(ctx context.Context, meta *types.NodeMeta, node *types.NodeState) error

	InsertPod(ctx context.Context, meta *types.NodeMeta, pod *types.PodNodeSpec) error
	UpdatePod(ctx context.Context, meta *types.NodeMeta, pod *types.PodNodeSpec) error
	RemovePod(ctx context.Context, meta *types.NodeMeta, pod *types.PodNodeSpec) error

	Remove(ctx context.Context, meta *types.NodeMeta) error
}