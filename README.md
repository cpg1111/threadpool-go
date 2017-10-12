# threadpool-go

A Threadpool of OS threads in Golang

## Why

While this library does defeat the purpose of one of Go's key feautres (goroutines), there are time in Go programs that cannot afford the amount of context switches that can come with a goroutine. And while goroutines will run on an OS thread until completion with some exceptions (those exceptions often being net.Conn's select/epoll IO or blocking on channel operations), there are times where bypassing the Go runtime's scheduler can be beneficial.

This library leverages that by providing an already instantiated pool of OS threads with more exposed control of what gets executed onto them.

