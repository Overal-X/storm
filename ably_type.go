package main

type AblyPublishArgs struct {
	Channel string
	Route   string
	Content interface{}
}

type AblySubscribeArgs struct {
	Channel  string
	Route    string
	Callback func(interface{}) error
}
