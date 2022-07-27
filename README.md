# crabviz-go

crabviz-go is a static analysis tool that generate call graph for Go projects, trying to make it easier to read the source code and the call graph itself.

## Preview

![preview](https://user-images.githubusercontent.com/20551552/181170301-f32bd74e-d48e-469e-8abc-1a5438eef659.gif)

## Usage

just run `crabviz-go <target package>`

## Features

* draw the outline of files
* group files by package
* highlight edges with different colors depending on calling relationships

## TODO

- [ ] interface implementation analysis
- [ ] make UI prettier
- [ ] add a sidebar

## Credits

crabviz-go is inspired by [graphql-voyager](https://github.com/APIs-guru/graphql-voyager) and [go-callvis](https://github.com/ofabry/go-callvis)
