package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	in "registry_service/internal/app/interfaces"
	"strconv"
	"time"

	"gopkg.in/validator.v2"
)

// @title Registry service
// @version 1.0
// @description Service responsible for register and manage order requests.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @Summary Create order entrypoint
// @Description Create order entrypoint
// @Produce json
// @Tags	orders
// @Success 200 {object} CreateOrderResponse
// @Failure 400 {object} ErrResponseMsg
// @Failure 500 {string} error
// @Param order body CreateOrderRequest true "order data"
// @Router /orders [POST]
func (s *Server) CreateOrder() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var orderData CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&orderData); err != nil {
			switch {
			case err == io.EOF:
				msg := ErrResponseMsg{Message: "Empty body"}
				JSONResponse(w, msg, http.StatusBadRequest)

				return
			case err != nil:
				msg := ErrResponseMsg{Message: err.Error()}
				JSONResponse(w, msg, http.StatusBadRequest)

				return
			}
		}

		if errs := validator.Validate(orderData); errs != nil {
			msg := ErrResponseMsg{Message: "Empty body"}
			JSONResponse(w, msg, http.StatusBadRequest)

			return
		}

		var orderItems []*in.MakeOrderItemDTO
		for _, v := range orderData.OrderItems {
			orderItems = append(orderItems, &in.MakeOrderItemDTO{
				ProductID: v.ProductID,
				Count:     v.Count,
			})
		}

		makeOrderData := &in.MakeOrderDTO{
			UserID:     orderData.UserID,
			OrderItems: orderItems,
		}

		err := s.App.OrdersService.MakeOrder(r.Context(), *makeOrderData)
		if err != nil {
			JSONResponse(w, err.Error(), http.StatusBadRequest)

			return
		}

		response := CreateOrderResponse{Status: "order request accepted"}

		JSONResponse(w, response, http.StatusOK)
	}

	return http.HandlerFunc(handler)
}

// @Summary List orders
// @Description List user orders
// @Produce json
// @Tags	orders
// @Success 200 {array} OrdersListResponse
// @Failure 400 {object} ErrResponseMsg
// @Failure 500 {string} error
// @Param user_id query CreateOrderRequest true "user id"
// @Router /orders [GET]
func (s *Server) OrderList() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.FormValue("user_id")

		userID, err := strconv.Atoi(userIDStr)
		if userIDStr == "" || err != nil || userID <= 0 {
			msg := ErrResponseMsg{Message: "user_id query param is not correct"}
			JSONResponse(w, msg, http.StatusBadRequest)

			return
		}

		orders, err := s.App.OrdersService.GetOrdersList(r.Context(), uint(userID))
		if err != nil {
			JSONResponse(w, err.Error(), http.StatusBadRequest)
		}

		ordersReponse := make([]OrdersListResponse, 0, 10)
		for _, v := range orders {
			ordersReponse = append(ordersReponse, OrdersListResponse{
				ID:             v.ID,
				UserID:         v.UserID,
				CreatedAt:      v.CreatedAt,
				Status:         v.Status,
				RejectedReason: v.RejectedReason,
			})
		}

		JSONResponse(w, ordersReponse, http.StatusOK)
	}

	return http.HandlerFunc(handler)
}

// @Summary List products
// @Description List products
// @Produce json
// @Tags	orders
// @Success 200 {array} ProductsListResponse
// @Failure 400 {object} ErrResponseMsg
// @Failure 500 {string} error
// @Router /products [GET]
func (s *Server) ProductsList() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		products, err := s.App.OrdersService.GetProductList(r.Context())
		if err != nil {
			JSONResponse(w, err.Error(), http.StatusBadRequest)
		}

		productsReponse := make([]ProductsListResponse, 0, 10)
		for _, v := range products {
			productsReponse = append(productsReponse, ProductsListResponse{
				ID:    v.ID,
				Title: v.Title,
				Price: v.Price,
			})
		}

		JSONResponse(w, productsReponse, http.StatusOK)
	}

	return http.HandlerFunc(handler)
}

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

		ordersDAOHealth := errToStr(s.App.OrdersDAO.HealthCheck(ctx))
		orderItemsDAOHealth := errToStr(s.App.OrderItemsDAO.HealthCheck(ctx))
		productPricesDAOHealth := errToStr(s.App.ProductPricesDAO.HealthCheck(ctx))
		brokerClientHealth := errToStr(s.App.BrokerClient.HealthCheck(ctx))

		response := HealthCheckResposne{
			OrdersConn:        ordersDAOHealth,
			OrderItemsConn:    orderItemsDAOHealth,
			ProductPricesConn: productPricesDAOHealth,
			BrokerConn:        brokerClientHealth,
		}

		JSONResponse(w, response, http.StatusOK)
	}

	return http.HandlerFunc(handler)
}
