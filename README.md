# Bollywood

Bollywood a super simplistic actor model framework. I wrote this to get an idea how 
an actor model could be implemented in Go. It is not meant to be used in production and doesn't
do support networking, limiting the actors to work within a single process.


## Usage



```go
    e := NewEngine()
    // spawn a baker actor 
	err := e.Spawn("baker", &baker{})
	if err != nil {
		panic(err)
	}
	// send a message to the baker
	err = e.Send("baker", &bakeBread{})
	if err != nil {
        panic(err)
    }
    // stop the baker:
    err = e.Stop("baker")
    if err != nil {
        panic(err)
    }

```
See the tests for usage



