package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "registry_service/docs"
	"registry_service/internal/app/registry"
	"registry_service/internal/pkg/log"
	"sync"

	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	App  *registry.App
	Serv *http.Server
}

func NewServer(app *registry.App) *Server {
	s := &Server{
		App: app,
	}

	router := createRouter(s)
	handler := log.LoggingMiddleware(app.Logger)(router)

	server := &http.Server{
		Addr:    app.Config.ServerAddr(),
		Handler: handler,
	}

	s.Serv = server

	return s
}

func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go s.App.OrdersService.EventPipeProcessor(ctx, &wg)
	go s.App.OrdersService.ConsumeRejectedOrderMsgLoop(ctx, &wg)
	go s.App.OrdersService.ConsumeSuccessMsgLoop(ctx, &wg)

	if err := s.Serv.ListenAndServe(); err != nil {
		cancel()
		wg.Wait()

		return err
	}

	cancel()
	wg.Wait()

	return nil
}

func (s *Server) Shutdown() {
	s.Serv.Close()
	s.App.Close()
}

func createRouter(s *Server) *mux.Router {
	r := mux.NewRouter()
	r.Handle("/health", s.HealthCheck()).Queries("timeout", "{[0-9]*?}").Methods(http.MethodGet)
	r.Handle("/health", s.HealthCheck()).Methods(http.MethodGet)
	r.Handle("/orders", s.CreateOrder()).Methods(http.MethodPost)
	r.Handle("/orders", s.OrderList()).Queries("user_id", "{[0-9]*?}").Methods(http.MethodGet)
	r.Handle("/products", s.ProductsList()).Methods(http.MethodGet)

	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", s.App.Config.ServerAddr())), // The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))

	return r
}

func JSONResponse(w http.ResponseWriter, msg interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err := json.NewEncoder(w).Encode(msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(code)
}
