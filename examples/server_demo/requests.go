package main

type LoginRequest struct {
	PlayerId string
	Name     string
}

type AddGoldRequest struct {
	PlayerId string
	Amount   int64
}

type RenameRequest struct {
	PlayerId string
	Name     string
}
