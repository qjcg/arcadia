title: Building Web Applications in Go
author:
  name: John Gosset
  url: https://www.ardanlabs.com
output: index.html
controls: true
theme: select/cleaver-select-theme
style: style.css

--

# Building Web Applications in Go
## QCon SF 2017 / Ardan Labs / John Gosset

https://qjcg.github.io/talks/qcon17/

--

# Introduction

--

### John Gosset

- developer and consultant
- Montreal, Canada
- writing Go since the language's early days
    - Came to Go from the [Plan9](https://9p.io/plan9/) community!
- work with organizations large and small
    - Canadian government
    - Fintech companies
    - Small carpet manufacturers in Saint-Jean-sur-Richelieu, Quebec
- Instructor with [Ardan Labs](https://www.ardanlabs.com/)
- Contractor for [RedHat](https://www.redhat.com/)
- Volunteer for the [Software Carpentry Foundation](https://software-carpentry.org/)

--

### Ardan Labs

> Trusted by Start-ups to Fortune 500 Companies, Ardan Labs is a high
> performance software development firm that builds and delivers reliable
> solutions and applications since 2010.

- https://www.ardanlabs.com/
- Go training, consulting, data science, design

--

### Workshop Outline

- HTTP Basics
- Testing
- Templating and Asset Management
- Muxing/Routing
- Serialization
- Middleware
- APIs
- TLS
- Auth

--

### Workshop Focus

- Server-side Web fundamentals, not:
    - front-end
    - frameworks (ex: [awesome-go: web frameworks](https://github.com/avelino/awesome-go#web-frameworks))
    - RPC (ex: [gRPC](https://grpc.io/))
    - etc

--

### Go Environment

- [Installing Go and your Workspace](https://www.goinggo.net/2016/05/installing-go-and-your-workspace.html)

To get the code and exercises for today's workshop:
    
```
go get github.com/qjcg/gotraining
```

**OPTIONAL** -- run the web app:

```
cd $GOPATH/src/github.com/qjcg/gotraining
go run main.go
```

Once the app is running, open your browser to: http://localhost:8080/
