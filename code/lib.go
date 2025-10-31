package main

import (
	"log"
	"strconv"
)

func sliceStrToIntConvert(slice []string) []int {
	var sliceNew []int
	for _, v := range slice {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Println(err)
			continue
		}
		sliceNew = append(sliceNew, n)
	}
	return sliceNew
}
