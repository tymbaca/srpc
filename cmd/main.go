package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func main() {
	s := &ServerImpl{}
	_ = s

	runreflect(s)
}

// WARN: T must be an interface
func runreflect[T any](s T) {
	// t := reflect.TypeFor[MyService]()
	t := reflect.TypeFor[T]()
	name := t.Name()
	if name == "" {
		name = t.Elem().Name()
		fmt.Println(`t.Name() was ""`)
	}
	fmt.Printf("t.Name(): %v\n", name)
	// fmt.Printf("t.Name(): %v\n", t.Name())
	// for i := range t.NumMethod() {
	// 	m := t.Method(i)
	// 	fmt.Printf("t.Method(%d): %v\n", i, m)
	// 	fmt.Printf("t.Method(%d) type: %v\n", i, m.Type)
	// }
	//
	v := reflect.ValueOf(s)
	//
	for i := range v.NumMethod() {
		m := v.Method(i)
		fmt.Printf("v.Method(%d).Type().Name(): %v\n", i, m.Type().Name())
		// fmt.Printf("v.Method(%d) type: %v\n", i, m.Type())
	}

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
