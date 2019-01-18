## 1 cover
```
$ cd $GOPATH/src/github.com/lambda-zhang/systemmonitor/
$ go test -coverprofile=covprofile
$ go tool cover -html=covprofile -o coverage.html
```

##
```
$ sudo apt-get install graphviz

<<<<<<< HEAD
$ go test -bench=".*" -cpuprofile=cprof -cpu=8 -memprofile=mprof
$ go tool pprof --text  cprof
$ go tool pprof --text  mprof
or
$ go tool pprof --web  cprof
$ go tool pprof --web  mprof
=======
$ go test -bench=".*" -cpuprofile=cprof
$ go tool pprof --text  cprof
or
$ go tool pprof --web  cprof

>>>>>>> f1cc068b545a35ab118aa5953118e3a3dc7537dd
```
