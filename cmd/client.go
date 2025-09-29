package main

import "github.com/tymbaca/srpc"

func genericclient() {
	c := srpc.NewClient[MyService]()
	c.Inner.GetNodes()
}
