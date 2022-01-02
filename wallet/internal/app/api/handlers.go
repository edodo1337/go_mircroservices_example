package api

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// @title Wallet service
// @version 1.0
// @description Service responsible for register and manage order requests.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @Summary Healthcheck
// @Description Check DB and broker client connections
// @Produce json
// @Tags	ops
// @Success 200 {object} HealthCheckResposne
// @Failure 400 {object} ErrResponseMsg
// @Failure 500 {string} error
// @Router /health [GET]
func (s *Server) HealthCheck() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		timeout := 15

		timeoutStr := r.FormValue("timeout")

		t, err := strconv.Atoi(timeoutStr)
		if timeoutStr != "" && err == nil {
			timeout = t
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeout)*time.Second)
		defer cancel()

		errToStr := func(err error) string {
			if err != nil {
				return err.Error()
			}

			return "ok"
		}

		walletsDAOHealth := errToStr(s.App.WalletsDAO.HealthCheck(ctx))
		walletTransDAOHealth := errToStr(s.App.WalletTransactionsDAO.HealthCheck(ctx))
		brokerClientHealth := errToStr(s.App.BrokerClient.HealthCheck(ctx))

		response := HealthCheckResposne{
			WalletsConn:     walletsDAOHealth,
			WalletTransConn: walletTransDAOHealth,
			BrokerConn:      brokerClientHealth,
		}

		JSONResponse(w, response, http.StatusOK)
	}

	return http.HandlerFunc(handler)
}
