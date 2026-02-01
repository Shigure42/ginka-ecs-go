package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

type Server struct {
	world   ginka_ecs_go.World
	auth    *AuthSystem
	wallet  *WalletSystem
	profile *ProfileSystem
}

func NewServer(world ginka_ecs_go.World, auth *AuthSystem, wallet *WalletSystem, profile *ProfileSystem) *Server {
	return &Server{world: world, auth: auth, wallet: wallet, profile: profile}
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
		PlayerId string `json:"player_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerId == "" || req.Name == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	if err := s.auth.Login(r.Context(), s.world, LoginRequest{PlayerId: req.PlayerId, Name: req.Name}); err != nil {
		http.Error(w, fmt.Sprintf("login: %v", err), http.StatusBadRequest)
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
		PlayerId string `json:"player_id"`
		Amount   int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerId == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	if err := s.wallet.AddGold(r.Context(), s.world, AddGoldRequest{PlayerId: req.PlayerId, Amount: req.Amount}); err != nil {
		http.Error(w, fmt.Sprintf("add gold: %v", err), http.StatusBadRequest)
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
		PlayerId string `json:"player_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerId == "" || req.Name == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	if err := s.profile.Rename(r.Context(), s.world, RenameRequest{PlayerId: req.PlayerId, Name: req.Name}); err != nil {
		http.Error(w, fmt.Sprintf("rename: %v", err), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
