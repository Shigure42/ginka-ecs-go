package main

type LoginCommand struct {
	PlayerId uint64
	Name     string
}

type AddGoldCommand struct {
	PlayerId uint64
	Amount   int64
}

type RenameCommand struct {
	PlayerId uint64
	Name     string
}
