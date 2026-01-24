package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

type Server struct {
	world ginka_ecs_go.World
}

func NewServer(world ginka_ecs_go.World) *Server {
	return &Server{world: world}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/add-gold", s.handleAddGold)
	mux.HandleFunc("/rename", s.handleRename)
	return mux
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PlayerId uint64 `json:"player_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerId == 0 || req.Name == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	if err := s.world.Submit(r.Context(), LoginCommand{PlayerId: req.PlayerId, Name: req.Name}); err != nil {
		http.Error(w, fmt.Sprintf("submit login: %v", err), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAddGold(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PlayerId uint64 `json:"player_id"`
		Amount   int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerId == 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	if err := s.world.Submit(r.Context(), AddGoldCommand{PlayerId: req.PlayerId, Amount: req.Amount}); err != nil {
		http.Error(w, fmt.Sprintf("submit add gold: %v", err), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PlayerId uint64 `json:"player_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerId == 0 || req.Name == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	if err := s.world.Submit(r.Context(), RenameCommand{PlayerId: req.PlayerId, Name: req.Name}); err != nil {
		http.Error(w, fmt.Sprintf("submit rename: %v", err), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
