package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pradiptaagus/go-chat-app/internal/model/domain"
	"github.com/pradiptaagus/go-chat-app/internal/model/web"
	"github.com/pradiptaagus/go-chat-app/pkg/util"
)

type ClientController interface {
	Create(ctx context.Context, w http.ResponseWriter, r *http.Request)
	FindAll(ctx context.Context, w http.ResponseWriter, r *http.Request)
}

type clientController struct {
	clients []*domain.Client
}

func NewClientController() ClientController {
	return &clientController{
		clients: make([]*domain.Client, 0),
	}
}

func (c *clientController) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	select {
	case <-ctx.Done():
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		return
	default:
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Parse the JSON body into CreateClientRequest struct
		var req web.CreateClientRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Create new client domain model
		client := &domain.Client{
			ID:          util.GenerateRandomID(),
			Username:    req.UserName,
			Name:        req.Name,
			CreatedDate: time.Now(),
		}

		// Save client to the controller's list
		c.clients = append(c.clients, client)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resBody := web.CreateClientResponse{
			UserID:      client.ID,
			Username:    client.Username,
			Name:        client.Name,
			CreatedDate: client.CreatedDate,
		}

		strResBody, err := json.Marshal(resBody)
		if err != nil {
			http.Error(w, "Failed to marshal response body", http.StatusInternalServerError)
			return
		}

		res := fmt.Sprintf(`{"status":"success","data": %s}`, strResBody)
		w.Write([]byte(res))
	}
}

func (c *clientController) FindAll(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	select {
	case <-ctx.Done():
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		return
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resBody := make([]web.CreateClientResponse, 0)
		for _, client := range c.clients {
			resBody = append(resBody, web.CreateClientResponse{
				UserID:      client.ID,
				Username:    client.Username,
				Name:        client.Name,
				CreatedDate: client.CreatedDate,
			})
		}

		strResBody, err := json.Marshal(resBody)
		if err != nil {
			http.Error(w, "Failed to marshal response body", http.StatusInternalServerError)
			return
		}

		res := fmt.Sprintf(`{"status":"success","data": %s}`, strResBody)
		w.Write([]byte(res))
	}
}
