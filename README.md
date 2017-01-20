# singular
The singular is an HTTP web framework based on httprouter.

PS: The warehouse is perfecting stage

##Use
	
	go get -u github.com/noaway/singular

##example

	package main

	import (	
		"fmt"
		"github.com/noaway/singular"
		"net/http"
	)

	type Student struct {
		Id   int
		Name string
		Age  int
	}

	func (stu *Student) Show() {
		fmt.Println(stu.Id, stu.Name, stu.Age)
	}

	func main() {
		r := singular.New()
		// r.Use(PlainText)

		v1 := r.Group("/v1")
		v1.Map(&Student{Id: 1, Name: "1234567890", Age: 25})
		v1.GET("/wang/:wy", func(ctx *singular.Context) {
			fmt.Println(ctx.Handles)
			ctx.Writer.Write([]byte("1"))
		}, PlainText)

		v1.POST("/wang/:wy", func(ctx *singular.Context) {
			fmt.Println(ctx.Handles)
			ctx.Writer.Write([]byte("12222"))
		})

		http.ListenAndServe(":8889", r)
	}

	func PlainText(f singular.Handle) singular.Handle {
		return func(ctx *singular.Context) {
			ctx.Set("qq", "1234567890")
			wy := ctx.Params.ByName("wy")
			if wy == "wy" {
				f(ctx)
			} else {
				ctx.Writer.Write([]byte("Parameter is not wy"))
			}
		}
	}


	
	
	
	