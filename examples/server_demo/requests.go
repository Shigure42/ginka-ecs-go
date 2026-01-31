package main

type LoginRequest struct {
	PlayerId uint64
	Name     string
}

type AddGoldRequest struct {
	PlayerId uint64
	Amount   int64
}

type RenameRequest struct {
	PlayerId uint64
	Name     string
}
