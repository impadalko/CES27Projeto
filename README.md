# CES27Projeto

We will be using the stardand Go programming setup as described in
https://golang.org/doc/code.html

## Setup environment

The environment variable %GOPATH% should point to the folder "~/go" in user dir.

```
mkdir -p "%GOPATH%/src/github.com/impadalko"
cd "%GOPATH%/src/github.com/impadalko"
git clone https://github.com/impadalko/CES27Projeto
cd CES27Projeto
```

## Building and running

```
cd "%GOPATH%/src/github.com/impadalko/CES27Projeto"
go build && ./CES27Projeto
```