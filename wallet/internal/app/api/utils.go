package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "wallet_service/docs"
	"wallet_service/internal/app/wallet"
	"wallet_service/internal/pkg/log"

	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	App  *wallet.App
	Serv *http.Server
}

func NewServer(app *wallet.App) *Server {
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
	if err := s.Serv.ListenAndServe(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go s.App.OrdersService.ConsumeRejectedOrderMsgLoop(ctx)

	return nil
}

func (s *Server) Shutdown() {
	s.Serv.Close()
}

func createRouter(s *Server) *mux.Router {
	r := mux.NewRouter()
	r.Handle("/health", s.HealthCheck()).Queries("timeout", "{[0-9]*?}").Methods(http.MethodGet)
	r.Handle("/orders", s.CreateOrder()).Methods(http.MethodPost)
	r.Handle("/orders", s.OrderList()).Queries("user_id", "{[0-9]*?}").Methods(http.MethodGet)

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
