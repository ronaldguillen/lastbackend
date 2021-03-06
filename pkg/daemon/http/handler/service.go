package handler

import (
	"encoding/json"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	e "github.com/lastbackend/lastbackend/libs/errors"
	"github.com/lastbackend/lastbackend/libs/model"
	c "github.com/lastbackend/lastbackend/pkg/daemon/context"
	"github.com/lastbackend/lastbackend/pkg/service"
	"github.com/lastbackend/lastbackend/pkg/util/generator"
	"io"
	"io/ioutil"
	"net/http"
)

func ServiceListH(w http.ResponseWriter, r *http.Request) {

	var (
		err          error
		session      *model.Session
		projectModel *model.Project
		ctx          = c.Get()
		params       = mux.Vars(r)
		projectParam = params["project"]
	)

	ctx.Log.Debug("List service handler")

	s, ok := context.GetOk(r, `session`)
	if !ok {
		ctx.Log.Error("Error: get session context")
		e.New("user").Unauthorized().Http(w)
		return
	}

	session = s.(*model.Session)

	projectModel, err = ctx.Storage.Project().GetByNameOrID(session.Uid, projectParam)
	if err == nil && projectModel == nil {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find project by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel, err := ctx.Storage.Service().ListByProject(session.Uid, projectModel.ID)
	if err != nil {
		ctx.Log.Error("Error: find services by user", err)
		e.HTTP.InternalServerError(w)
		return
	}

	servicesSpec, err := service.List(ctx.K8S, projectModel.ID)
	if err != nil {
		ctx.Log.Error("Error: get serivce spec from cluster", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	var list = model.ServiceList{}
	var response []byte

	if serviceModel != nil {
		for _, val := range *serviceModel {
			val.Spec = servicesSpec["lb-"+val.ID]
			list = append(list, val)
		}

		response, err = list.ToJson()
		if err != nil {
			ctx.Log.Error("Error: convert struct to json", err.Error())
			e.HTTP.InternalServerError(w)
			return
		}

	} else {
		response, err = serviceModel.ToJson()
		if err != nil {
			ctx.Log.Error("Error: convert struct to json", err.Error())
			e.HTTP.InternalServerError(w)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		ctx.Log.Error("Error: write response", err.Error())
		return
	}
}

func ServiceInfoH(w http.ResponseWriter, r *http.Request) {

	var (
		err          error
		session      *model.Session
		projectModel *model.Project
		serviceModel *model.Service
		ctx          = c.Get()
		params       = mux.Vars(r)
		projectParam = params["project"]
		serviceParam = params["service"]
	)

	ctx.Log.Debug("Get service handler")

	s, ok := context.GetOk(r, `session`)
	if !ok {
		ctx.Log.Error("Error: get session context")
		e.New("user").Unauthorized().Http(w)
		return
	}

	session = s.(*model.Session)

	projectModel, err = ctx.Storage.Project().GetByNameOrID(session.Uid, projectParam)
	if err == nil && projectModel == nil {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find project by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel, err = ctx.Storage.Service().GetByNameOrID(session.Uid, serviceParam)
	if err == nil && (serviceModel == nil || serviceModel.Project != projectModel.ID) {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find service by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceSpec, err := service.Get(ctx.K8S, serviceModel.Project, "lb-"+serviceModel.ID)
	if err != nil {
		ctx.Log.Error("Error: get serivce spec from cluster", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel.Spec = serviceSpec

	response, err := serviceModel.ToJson()
	if err != nil {
		ctx.Log.Error("Error: convert struct to json", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		ctx.Log.Error("Error: write response", err.Error())
		return
	}
}

type serviceUpdateS struct {
	*model.ServiceUpdateConfig
}

func (s *serviceUpdateS) decodeAndValidate(reader io.Reader) *e.Err {

	var (
		err error
		ctx = c.Get()
	)

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		ctx.Log.Error(err)
		return e.New("user").Unknown(err)
	}

	err = json.Unmarshal(body, s)
	if err != nil {
		return e.New("service").IncorrectJSON(err)
	}

	return nil
}

func ServiceUpdateH(w http.ResponseWriter, r *http.Request) {

	var (
		err          error
		session      *model.Session
		projectModel *model.Project
		serviceModel *model.Service
		ctx          = c.Get()
		params       = mux.Vars(r)
		projectParam = params["project"]
		serviceParam = params["service"]
	)

	ctx.Log.Debug("Update service handler")

	s, ok := context.GetOk(r, `session`)
	if !ok {
		ctx.Log.Error("Error: get session context")
		e.New("user").Unauthorized().Http(w)
		return
	}

	session = s.(*model.Session)

	// request body struct
	rq := new(serviceUpdateS)
	if err := rq.decodeAndValidate(r.Body); err != nil {
		ctx.Log.Error("Error: validation incomming data", err)
		err.Http(w)
		return
	}

	projectModel, err = ctx.Storage.Project().GetByNameOrID(session.Uid, projectParam)
	if err == nil && projectModel == nil {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find project by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel, err = ctx.Storage.Service().GetByNameOrID(session.Uid, serviceParam)
	if err == nil && (serviceModel == nil || serviceModel.Project != projectModel.ID) {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find service by name or id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	if rq.Name != nil {
		serviceModel.Name = *rq.Name
	}

	if rq.Description != nil {
		serviceModel.Description = *rq.Description
	}
	serviceModel, err = ctx.Storage.Service().Update(serviceModel)
	if err != nil {
		ctx.Log.Error("Error: insert service to db", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	cfg := rq.CreateServiceConfig()

	err = service.Update(ctx.K8S, serviceModel.Project, "lb-"+serviceModel.ID, cfg)
	if err != nil {
		ctx.Log.Error("Error: update service", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceSpec, err := service.Get(ctx.K8S, serviceModel.Project, "lb-"+serviceModel.ID)
	if err != nil {
		ctx.Log.Error("Error: get serivce spec from cluster", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel.Spec = serviceSpec

	response, err := serviceModel.ToJson()
	if err != nil {
		ctx.Log.Error("Error: convert struct to json", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		ctx.Log.Error("Error: write response", err.Error())
		return
	}
}

func ServiceRemoveH(w http.ResponseWriter, r *http.Request) {

	var (
		er           error
		ctx          = c.Get()
		session      *model.Session
		projectModel *model.Project
		params       = mux.Vars(r)
		projectParam = params["project"]
		serviceParam = params["service"]
	)

	ctx.Log.Info("Remove service")

	s, ok := context.GetOk(r, `session`)
	if !ok {
		ctx.Log.Error("Error: get session context")
		e.New("user").Unauthorized().Http(w)
		return
	}

	session = s.(*model.Session)

	projectModel, err := ctx.Storage.Project().GetByNameOrID(session.Uid, projectParam)
	if err == nil && projectModel == nil {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find project by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel, err := ctx.Storage.Service().GetByNameOrID(session.Uid, serviceParam)
	if err == nil && (serviceModel == nil || serviceModel.Project != projectModel.ID) {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find service by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	err = ctx.Storage.Hook().RemoveByService(serviceModel.ID)
	if err != nil {
		ctx.Log.Error("Error: remove hook from db", err)
		e.HTTP.InternalServerError(w)
		return
	}

	err = ctx.Storage.Activity().RemoveByService(session.Uid, serviceModel.ID)
	if err != nil {
		ctx.Log.Error("Error: remove activity from db", err)
		e.HTTP.InternalServerError(w)
		return
	}

	err = service.Remove(ctx.K8S, serviceModel.Project, "lb-"+serviceModel.ID)
	if err != nil {
		ctx.Log.Error("Error: remove service from kubernetes", err)
		e.HTTP.InternalServerError(w)
		return
	}

	err = ctx.Storage.Service().Remove(session.Uid, serviceModel.ID)
	if err != nil {
		ctx.Log.Error("Error: remove service from db", err)
		e.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, er = w.Write([]byte{})
	if er != nil {
		ctx.Log.Error("Error: write response", er.Error())
		return
	}
}

func ServiceActivityListH(w http.ResponseWriter, r *http.Request) {

	var (
		err          error
		session      *model.Session
		projectModel *model.Project
		ctx          = c.Get()
		params       = mux.Vars(r)
		projectParam = params["project"]
		serviceParam = params["service"]
	)

	ctx.Log.Debug("List service activity handler")

	s, ok := context.GetOk(r, `session`)
	if !ok {
		ctx.Log.Error("Error: get session context")
		e.New("user").Unauthorized().Http(w)
		return
	}

	session = s.(*model.Session)

	projectModel, err = ctx.Storage.Project().GetByNameOrID(session.Uid, projectParam)
	if err == nil && projectModel == nil {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find project by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	activityListModel, err := ctx.Storage.Activity().ListServiceActivity(session.Uid, serviceParam)
	if err != nil {
		ctx.Log.Error("Error: find service avtivity list by id", err)
		e.HTTP.InternalServerError(w)
		return
	}

	response, err := activityListModel.ToJson()
	if err != nil {
		ctx.Log.Error("Error: convert struct to json", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		ctx.Log.Error("Error: write response", err.Error())
		return
	}
}

func ServiceLogsH(w http.ResponseWriter, r *http.Request) {

	var (
		ctx            = c.Get()
		session        *model.Session
		projectModel   *model.Project
		serviceModel   *model.Service
		params         = mux.Vars(r)
		projectParam   = params["project"]
		serviceParam   = params["service"]
		query          = r.URL.Query()
		podQuery       = query.Get("pod")
		containerQuery = query.Get("container")
		ch             = make(chan bool, 1)
		notify         = w.(http.CloseNotifier).CloseNotify()
	)

	ctx.Log.Info("Show service log")

	go func() {
		<-notify
		ctx.Log.Debug("HTTP connection just closed.")
		ch <- true
	}()

	s, ok := context.GetOk(r, `session`)
	if !ok {
		ctx.Log.Error("Error: get session context")
		e.New("user").Unauthorized().Http(w)
		return
	}

	session = s.(*model.Session)

	projectModel, err := ctx.Storage.Project().GetByNameOrID(session.Uid, projectParam)
	if err == nil && projectModel == nil {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find project by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel, err = ctx.Storage.Service().GetByNameOrID(session.Uid, serviceParam)
	if err == nil && (serviceModel == nil || serviceModel.Project != projectModel.ID) {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find service by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	opts := service.ServiceLogsOption{
		Stream:     w,
		Pod:        podQuery,
		Container:  containerQuery,
		Follow:     true,
		Timestamps: true,
	}

	service.Logs(ctx.K8S, serviceModel.Project, &opts, ch)
}

func ServiceHookCreateH(w http.ResponseWriter, r *http.Request) {

	var (
		err          error
		session      *model.Session
		projectModel *model.Project
		serviceModel *model.Service
		hookModel    *model.Hook
		ctx          = c.Get()
		params       = mux.Vars(r)
		projectParam = params["project"]
		serviceParam = params["service"]
	)

	ctx.Log.Debug("List hook create handler")

	s, ok := context.GetOk(r, `session`)
	if !ok {
		ctx.Log.Error("Error: get session context")
		e.New("user").Unauthorized().Http(w)
		return
	}

	session = s.(*model.Session)

	projectModel, err = ctx.Storage.Project().GetByNameOrID(session.Uid, projectParam)
	if err == nil && projectModel == nil {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find project by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel, err = ctx.Storage.Service().GetByNameOrID(session.Uid, serviceParam)
	if err == nil && (serviceModel == nil || serviceModel.Project != projectModel.ID) {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find service by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	hookModel = &model.Hook{
		User:    serviceModel.User,
		Token:   generator.GenerateToken(32),
		Service: serviceModel.ID,
	}

	hookModel, err = ctx.Storage.Hook().Insert(hookModel)
	if err != nil {
		ctx.Log.Error("Error: find hook list by user", err)
		e.HTTP.InternalServerError(w)
		return
	}

	response, err := hookModel.ToJson()
	if err != nil {
		ctx.Log.Error("Error: convert struct to json", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		ctx.Log.Error("Error: write response", err.Error())
		return
	}
}

func ServiceHookListH(w http.ResponseWriter, r *http.Request) {

	var (
		err           error
		session       *model.Session
		projectModel  *model.Project
		serviceModel  *model.Service
		hookListModel *model.HookList
		ctx           = c.Get()
		params        = mux.Vars(r)
		projectParam  = params["project"]
		serviceParam  = params["service"]
	)

	ctx.Log.Debug("List hook handler")

	s, ok := context.GetOk(r, `session`)
	if !ok {
		ctx.Log.Error("Error: get session context")
		e.New("user").Unauthorized().Http(w)
		return
	}

	session = s.(*model.Session)

	projectModel, err = ctx.Storage.Project().GetByNameOrID(session.Uid, projectParam)
	if err == nil && projectModel == nil {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find project by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel, err = ctx.Storage.Service().GetByNameOrID(session.Uid, serviceParam)
	if err == nil && (serviceModel == nil || serviceModel.Project != projectModel.ID) {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find service by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	hookListModel, err = ctx.Storage.Hook().ListByService(session.Uid, serviceModel.ID)
	if err != nil {
		ctx.Log.Error("Error: find hook list by user", err)
		e.HTTP.InternalServerError(w)
		return
	}

	response, err := hookListModel.ToJson()
	if err != nil {
		ctx.Log.Error("Error: convert struct to json", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		ctx.Log.Error("Error: write response", err.Error())
		return
	}
}

func ServiceHookRemoveH(w http.ResponseWriter, r *http.Request) {

	var (
		err          error
		ctx          = c.Get()
		session      *model.Session
		projectModel *model.Project
		serviceModel *model.Service
		params       = mux.Vars(r)
		projectParam = params["project"]
		serviceParam = params["service"]
		hookParam    = params["hook"]
	)

	ctx.Log.Info("List service hook handler")

	s, ok := context.GetOk(r, `session`)
	if !ok {
		ctx.Log.Error("Error: get session context")
		e.New("user").Unauthorized().Http(w)
		return
	}

	session = s.(*model.Session)

	projectModel, err = ctx.Storage.Project().GetByNameOrID(session.Uid, projectParam)
	if err == nil && projectModel == nil {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find project by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	serviceModel, err = ctx.Storage.Service().GetByNameOrID(session.Uid, serviceParam)
	if err == nil && (serviceModel == nil || serviceModel.Project != projectModel.ID) {
		e.New("service").NotFound().Http(w)
		return
	}
	if err != nil {
		ctx.Log.Error("Error: find service by id", err.Error())
		e.HTTP.InternalServerError(w)
		return
	}

	err = ctx.Storage.Hook().Remove(hookParam)
	if err != nil {
		ctx.Log.Error("Error: remove hook from db", err)
		e.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte{})
	if err != nil {
		ctx.Log.Error("Error: write response", err.Error())
		return
	}
}
