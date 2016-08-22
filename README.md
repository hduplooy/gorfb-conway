# gorfb-examples

This repository have some examples how to make use of the [hduplooy/gorfb](https://github.com/hduplooy/gorfb) library.

## gorfb-conway.go 

A [Conway's Game of Life](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life) implementation that runs as a RFB server at 1366x768 pixels of which only 700x700 pixels is used to view the board. The RFB server is started on the localhost at port 5901. When a connection is made the board is initialized with random information and then it is updated every half a second with a new board based on the rules of the standard rules. This is just a basic implementation with minimal comments for now. 

When the server is started open a VNC Client and point it to your localhost at port 5901. The server will then require authentication, use 'conway12' as the authentication text. The starting board will now be shown with a graph on the right. To start the simulation press the 'p' key, to pause it again press 'p' again. Left click on any cell to switch it on and right click on it to switch it off. 

As the simulation run the board will be updated as each iteration is processed and the graph on the right will be update to show the current population size.

Please note the program also makes use of [golang.org/x/image/font](https://godoc.org/golang.org/x/image/font) to draw text in the frame buffer so please install it into your GOPATH.


