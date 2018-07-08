# gracefulServer
gracefulServer is a simple library which makes it easy to graceful restart golang server。

# Usage
You can find the usage of this library from the [demo](./test/test.go)

Update server steps:
1. replace the server file or configure files
2. find the processor id of the server
3. exec `kill -USR2 [pid]`