package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func main() {
	var s MyService = &ServerImpl{}
	_ = s
	// t := reflect.TypeFor[MyService]()
	// t := reflect.TypeOf(s)
	// fmt.Printf("t.Name(): %v\n", t.Name())
	// for i := range t.NumMethod() {
	// 	m := t.Method(i)
	// 	fmt.Printf("t.Method(%d): %v\n", i, m)
	// 	fmt.Printf("t.Method(%d) type: %v\n", i, m.Type)
	// }
	//
	v := reflect.ValueOf(s)
	//
	// for i := range v.NumMethod() {
	// 	m := v.Method(i)
	// 	fmt.Printf("v.Method(%d): %v\n", i, m)
	// 	fmt.Printf("v.Method(%d) type: %v\n", i, m.Type())
	// }

	// fmt.Printf("v.MethodByName(\"GetNodes\").Type(): %v\n", v.MethodByName("GetNodes").Type())
	getNodesMethod := v.MethodByName("GetNodes")
	// respVals := getNodesMethod.Call(toValues(context.Background(), GetNodesReq{arg: 10}))
	// for i, v := range respVals {
	// 	fmt.Printf("respVals[%d]: %#v\n", i, v.Interface())
	// }

	// getNodesMethodType := getNodesMethod.Type()
	// for i := range getNodesMethodType.NumIn() {
	// 	inType := getNodesMethodType.In(i)
	// 	fmt.Printf("getNodesMethodType.In(i): %v\n", inType)
	// 	fmt.Printf("inType.PkgPath(): %v\n", inType.PkgPath())
	// 	reflect.New(inType)
	// }

	getNodesReqVal := reflect.New(getNodesMethod.Type().In(1))
	fmt.Printf("getNodesReqVal.Interface(): %#v\n", getNodesReqVal.Interface())
	err := json.Unmarshal([]byte(`{"Arg": 99}`), getNodesReqVal.Interface())
	if err != nil {
		panic(err)
	}

	fmt.Printf("getNodesReqVal.Interface(): %#v\n", getNodesReqVal.Interface())
}

func toValues(ins ...any) []reflect.Value {
	outs := make([]reflect.Value, 0, len(ins))
	for _, in := range ins {
		v := reflect.ValueOf(in)
		outs = append(outs, v)
	}

	return outs
}
