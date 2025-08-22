package main

import (
	"fmt"
	"time"
)

type usecase struct {
}

func NewUsecase() *usecase {
	return &usecase{}
}

func (u *usecase) printConf() {
	t := time.NewTicker(1 * time.Second)

	go func() {
		for range t.C {
			fmt.Println(GetConf())
		}
	}()
}
