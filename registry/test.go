package main

import (
	"encoding/json"
	"fmt"

	"github.com/creasty/defaults"
	"gopkg.in/validator.v2"
)

type Item struct {
	X int `default:"10"`
}

func main() {
	i := Item{}
	if err := defaults.Set(&i); err != nil {
		panic(err)
	}
	fmt.Println(i)

	type NewUserRequest struct {
		Username string `validate:"min=3,max=40,regexp=^[a-zA-Z]*$"`
		Name     string `validate:"nonzero"`
		Age      int    `validate:"min=21,nonzero"`
		Password string `validate:"min=8"`
		Items    []int  `validate:"min=2"`
	}

	nur := NewUserRequest{Items: []int{1}}

	val, err := json.Marshal(nur)
	if err != nil {
		fmt.Println(err, val)
		return
	}

	fmt.Println("->>>", err.Error())

	if errs := validator.Validate(nur); errs != nil {
		fmt.Println(errs.Error())
	}

	fmt.Println(nur)
}
