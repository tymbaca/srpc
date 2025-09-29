package main

import (
	"context"
	"fmt"
	"reflect"
)

func main() {
	var s MyService = &ServerImpl{}
	_ = s
	t := reflect.TypeFor[MyService]()
	fmt.Printf("t.Name(): %v\n", t.Name())
	for i := range t.NumMethod() {
		m := t.Method(i)
		fmt.Printf("v.Method(i): %v\n", m)
	}

	v := reflect.ValueOf(s)

	// fmt.Printf("v.MethodByName(\"GetNodes\").Type(): %v\n", v.MethodByName("GetNodes").Type())
	v.MethodByName("GetNodes").Call(toValues(context.Background(), GetNodesReq{arg: 10}))
}

func toValues(ins ...any) []reflect.Value {
	outs := make([]reflect.Value, 0, len(ins))
	for _, in := range ins {
		v := reflect.ValueOf(in)
		outs = append(outs, v)
	}

	return outs
}
