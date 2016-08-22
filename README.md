# gorfb-examples

This repository have some examples how to make use of the [hduplooy/gorfb](https://github.com/hduplooy/gorfb) library.

gorfb-conway.go is a [Conway's Game of Life](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life) implementation that runs as a RFB server at 1366x768 pixels of which only 700x700 pixels is used to view the board. The RFB server is started on the localhost at port 5901. When a connection is made the board is initialized with random information and then it is updated every half a second with a new board based on the rules of the standard rules. This is just a basic implementation with minimal comments for now. When started open a VNC Client and point it to your localhost at port 5901 and you will see the board and as it changes over time. Just disconnect to stop.

