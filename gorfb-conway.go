// gorfb-conway project main.go
package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"time"

	"github.com/hduplooy/gorfb"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type GOL struct { // Game Of Life data
	// Image where we build our frame buffer to view
	Img *image.RGBA
	// The board representation
	board [][]byte
	// Total population per increment to a max of 600 increments after that the totals are windowed
	tots []int
	// flag to indicate if the simulation is currently running
	running bool
	// The connection
	RFBConn *gorfb.RFBConn
}

// Some basic colors that we use
var black = color.RGBA{0, 0, 0, 0}
var darkgray = color.RGBA{128, 128, 128, 0}
var white = color.RGBA{255, 255, 255, 0}
var green = color.RGBA{128, 255, 128, 0}

// sendRectangle will take a rectangle of data from the image buffer and tell gorfb to send it to the client
func (gol *GOL) sendRectangle(x, y, width, height int) error {
	if x == 0 && y == 0 && width == 1366 && height == 768 { // if full screen then send whole buffer
		return gol.RFBConn.SendRectangle(0, 0, 1366, 768, gol.Img.Pix)
	}
	// Total number of bytes to send
	sz := width * height * 4
	buf := make([]byte, sz)
	if x == 0 && width == 1366 {
		// If the width and x fits the frame buffer then just copy the lines that need to be send
		copy(buf[0:], gol.Img.Pix[y*1366*4:(y+height-1)*1366*4])
	} else {
		// Start position in the image buffer
		st1 := (y*1366 + x) * 4
		// Start position in buffer we are going to send
		st2 := 0

		w4 := width * 4 // byte size of every line

		// for every line copy it from source to buffer
		for i := 0; i < height; i, st1, st2 = i+1, st1+1366*4, st2+w4 {
			copy(buf[st2:], gol.Img.Pix[st1:st1+w4])
		}
	}
	// Tell gorfb to send it
	return gol.RFBConn.SendRectangle(x, y, width, height, buf)
}

// DrawHLine will draw a horizontal line into a RGBA image buffer (of pixel depth 32)
// x1 to x2 horizontally at height y
func (gol *GOL) DrawHLine(x1, x2, y int, col color.RGBA) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	// If not on image area return
	if x2 < 0 || x1 >= 1366 || y < 0 || y >= 768 {
		return
	}
	// No negatives allowed
	if x1 < 0 {
		x1 = 0
	}
	// Cannot go beyond width
	if x2 >= 1366 {
		x2 = 1365
	}
	// Start position in image buffer
	strt := (y*1366 + x1) * 4
	// End position in image buffer
	end := strt + (x2-x1)*4
	// Set the pixels from start to end to the color provided
	for strt <= end {
		gol.Img.Pix[strt] = col.R
		gol.Img.Pix[strt+1] = col.G
		gol.Img.Pix[strt+2] = col.B
		strt += 4
	}
}

// DrawVLine will draw a vertical line into a RGBA image buffer (of pixel depth 32)
// y1 to y2 vertically at horizontal position x
func (gol *GOL) DrawVLine(x, y1, y2 int, col color.RGBA) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	// If not on image area return
	if x < 0 || x >= 1365 || y2 < 0 || y1 >= 767 {
		return
	}
	// No negatives allowed
	if y1 < 0 {
		y1 = 0
	}
	// Cannot go beyond height
	if y2 >= 768 {
		y2 = 767
	}
	// Start position in image buffer
	strt := (y1*1366 + x) * 4
	// End position in image buffer
	hght := y2 - y1
	// Set the pixels from start to end to the color provided
	for j := 0; j < hght; j++ {
		gol.Img.Pix[strt] = col.R
		gol.Img.Pix[strt+1] = col.G
		gol.Img.Pix[strt+2] = col.B
		strt += 1366 * 4 // buffer is 1366 pixels and 4 bytes per pixel
	}
}

// DrawRect draws a record with the specified bounds into a RGBA image buffer (of pixel depth 32)
// DrawHLine and DrawVLine is used to draw the lines
func (gol *GOL) DrawRect(x1, y1, x2, y2 int, col color.RGBA) {
	if y1 >= 0 && y1 < 768 {
		gol.DrawHLine(x1, x2, y1, col)
	}
	if y1 >= 0 && y1 < 768 {
		gol.DrawHLine(x1, x2, y2, col)
	}
	if x1 >= 0 && x1 < 1366 {
		gol.DrawVLine(x1, y1, y2, col)
	}
	if x2 >= 0 && x2 < 1366 {
		gol.DrawVLine(x2, y1, y2, col)
	}
}

// DrawRect draws a record with the specified bounds into a RGBA image buffer (of pixel depth 32)
func (gol *GOL) FillRect(x1, y1, x2, y2 int, col color.RGBA) {
	// swap x1 and x2 if x1 not less than x2
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	// swap y1 and y2 if y1 not less than y2
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	// if not on the area of the image buffer then return
	if x2 < 0 || x1 >= 1366 || y2 < 0 || y1 >= 768 {
		return
	}
	// Cannot go less than zero
	if x1 < 0 {
		x1 = 0
	}
	// Cannot go beyond width
	if x2 >= 1366 {
		x2 = 1365
	}
	// Cannot go less than zero
	if y1 < 0 {
		y1 = 0
	}
	// Cannot go beyond height
	if y2 >= 768 {
		y2 = 767
	}
	// Start position within image buffer
	strt := (y1*1366 + x1) * 4
	hght := y2 - y1 // Height of rectangle
	wdth := x2 - x1 // Width of rectangle
	for j := 0; j < hght; j++ {
		strt2 := strt // Position within buffer for going horizontal
		// Do the horizontal line of values
		for i := 0; i < wdth; i++ {
			gol.Img.Pix[strt2] = col.R
			gol.Img.Pix[strt2+1] = col.G
			gol.Img.Pix[strt2+2] = col.B
			strt2 += 4
		}
		// Now for the next line
		strt += 1366 * 4
	}
}

// Make use of the basic stuff from golang.org/x/image/font
func (gol GOL) DrawText(x, y int, val string) {
	// Create the drawer structure and tell it the image buffer where to draw, the color (we like black), the font face andthe position
	drw := &font.Drawer{Dst: gol.Img, Src: image.Black, Face: basicfont.Face7x13, Dot: fixed.P(x, y)}
	// Draw the string now
	drw.DrawString(val)
}

// Update the population graph on the image buffer
func (gol *GOL) UpdateGraph() {
	gol.DrawText(745, 108, "Population Size")
	// Clear area where to put the graph
	gol.FillRect(745, 110, 1345, 710, white)
	gol.DrawRect(745, 110, 1345, 710, black)

	// Draw the totals as green lines from bottom to height according to total size
	for i, val := range gol.tots {
		gol.DrawVLine(746+i, 709-val, 709, green)
	}

	// Now draw some check lines within the graph
	for i := 0; i < 600; i += 10 { // Only every 10 pixels
		if i%100 == 0 { // If it is the 100'th pixel
			gol.DrawVLine(745+i, 110, 710, darkgray)       // Draw it the full height
			gol.DrawHLine(745, 1345, 710-i, darkgray)      // Draw it the full width
			gol.DrawText(752, 708-i, fmt.Sprintf("%d", i)) // Draw text indicating the population size
		} else {
			gol.DrawVLine(745+i, 705, 710, black)
			gol.DrawHLine(745, 750, 710-i, black)
		}
	}
}

// Update the simulation board
func (gol *GOL) Update() {
	// Display individual cells of board in place
	gol.FillRect(0, 0, 1366, 768, white)
	gol.DrawText(745, 27, "Conway's Game of Life")
	gol.DrawHLine(745, 900, 29, black)
	gol.DrawText(745, 50, "Press \"p\" key to start or pause the simulation.")
	gol.DrawText(745, 70, "Left click on cells to activate and right click to de-activate.")

	// Draw the individual cells that are active
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			if gol.board[i][j] == 1 {
				gol.FillRect(10+i*7, 10+j*7, 17+i*7, 17+j*7, black)
			}
		}
	}
	// Draw lines
	for i := 0; i <= 700; i += 7 {
		gol.DrawHLine(10, 710, i+10, black)
		gol.DrawVLine(i+10, 10, 710, black)
	}
	gol.UpdateGraph()
}

// Update a single cell
func (gol *GOL) UpdateCell(i, j int) {
	if gol.board[i][j] == 1 {
		gol.FillRect(10+i*7, 10+j*7, 17+i*7, 17+j*7, black)
	} else {
		gol.FillRect(10+i*7, 10+j*7, 17+i*7, 17+j*7, white)
		gol.DrawRect(10+i*7, 10+j*7, 17+i*7, 17+j*7, black)
	}
}

// Check value for position provided adjust if it wraps over the board
func check(oldboard [][]byte, i, j int) byte {
	if i < 0 {
		i = 99
	}
	if j < 0 {
		j = 99
	}
	if i >= 100 {
		i = 0
	}
	if j >= 100 {
		j = 0
	}
	return oldboard[i][j]
}

// checkCnt check the neighbour count for a cell
// check will automatically adjust for board border positions to wrap around
func checkCnt(oldboard [][]byte, i, j int) int {
	cnt := 0
	for a := -1; a <= 1; a++ {
		for b := -1; b <= 1; b++ {
			if a != 0 || b != 0 {
				cnt += int(check(oldboard, i+a, j+b))
			}
		}
	}
	return cnt
}

// Run will execute the main loop to do the simulation
// Every .3 seconds the next iteration is processed
func (gol *GOL) Run() {
	// Make a backup of the current board to oldboard
	// then we can build a new board up from the old values
	var oldboard [][]byte = make([][]byte, 100)
	for i := 0; i < 100; i++ {
		oldboard[i] = make([]byte, 100)
	}
	for {

		<-time.After(time.Nanosecond * 300000000) // .3 Seconds for every step
		popcnt := 0
		if gol.running {
			// Make a copy of the board to work with the previous instance separate from the new one
			for i := 0; i < 100; i++ {
				copy(oldboard[i], gol.board[i])
			}
			// Process board acording to rules
			for i := 0; i < 100; i++ {
				for j := 0; j < 100; j++ {
					cnt := checkCnt(oldboard, i, j) // Determine how many neighbours
					if oldboard[i][j] == 0 {        // If cell is currently de-activates
						if cnt == 3 { // If de-activated and 3 neighbours then activate
							gol.board[i][j] = 1
						}
					} else {
						if cnt < 2 || cnt > 3 { // if activated and either less than 3 neighbours or more than 3 then de-activate
							gol.board[i][j] = 0
						}
					}
					if gol.board[i][j] != oldboard[i][j] {
						// Update the individual cell if it has changed state
						gol.UpdateCell(i, j)
					}
					// update population count for this iteration
					popcnt += int(gol.board[i][j])
				}
			}
			popcnt = (popcnt * 100) / 1667 // Adjust count to fit within 600 pixels
			if len(gol.tots) == 600 {
				// If gol.tots already have 600 entries
				// Just slide everything down one and put new one at the end
				copy(gol.tots, gol.tots[1:])
				gol.tots[599] = popcnt
			} else {
				// Else just append to the end
				gol.tots = append(gol.tots, popcnt)
			}
			// Send the updated board to the client
			if gol.sendRectangle(10, 10, 701, 701) != nil {
				break
			}
			gol.UpdateGraph()
			// Send the updated graph to the client
			if gol.sendRectangle(745, 210, 601, 501) != nil {
				break
			}

		}
	}
}

// Init init function as called be gofrb when a RFB connection is made
// A RGBA image is created where the view is build
// A board is initialized randomly
// The updated view is send to the client
// Then start the main routine as a go routine
func (gol *GOL) Init(conn *gorfb.RFBConn) {
	gol.RFBConn = conn
	gol.Img = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1366, 768}})
	gol.board = make([][]byte, 100)
	gol.tots = make([]int, 0, 600)
	for i := 0; i < 100; i++ {
		gol.board[i] = make([]byte, 100)
		for j := 0; j < 100; j++ {
			if rand.Intn(100) < 45 {
				gol.board[i][j] = 1
			}
		}
	}
	gol.Update()
	go gol.Run()
}

// ProcessSetPixelFormat as requested by client
// We ignore it and only print it out
func (gol *GOL) ProcessSetPixelFormat(conn *gorfb.RFBConn, pf gorfb.PixelFormat) {
	fmt.Printf("PixelFormat request=%t\n", pf)
}

// ProcessSetEncoding as requested by client
// It is ignored here
func (gol *GOL) ProcessSetEncoding(conn *gorfb.RFBConn, encodings []int) {
	// Ignored for now
}

// ProcessUpdateRequest as requested by client
// Only if it is not an incremental request do we honor it
func (gol *GOL) ProcessUpdateRequest(conn *gorfb.RFBConn, x, y, width, height int, incremental bool) {
	if !incremental { // Only send an update if it is not incremental
		gol.sendRectangle(x, y, width, height)
	}
}

// ProcessKeyEvent as requested by client
// key is the key that is pressed or released
// downflag is true if it is pressed or false when released
func (gol *GOL) ProcessKeyEvent(conn *gorfb.RFBConn, key int, downflag bool) {
	if key == 'p' && downflag { // We are only interrested in a 'p' when it is pressed down
		gol.running = !gol.running // Change state of running flag
	}
}

// ProcessPointerEvent as requested by client
// x,y are the coordinates
// button is the button bitmask for buttons pressed
func (gol *GOL) ProcessPointerEvent(conn *gorfb.RFBConn, x, y, button int) {
	// If the point is within the board on the image buffer
	if button >= 1 && x >= 10 && x < 710 && y >= 10 && y < 710 {
		// Calculate the board position
		i := (x - 10) / 7
		j := (y - 10) / 7
		if button == 1 {
			// If left button pressed then activate
			gol.board[i][j] = 1
		} else {
			// iF any other button pressed de-activate
			gol.board[i][j] = 0
		}
		// Update the individual cell
		gol.UpdateCell(i, j)
		if !gol.running {
			// If the simulation is not running at the moment send the update to the client
			// else it will automatically be shown when the simulation update
			gol.sendRectangle(10+i*7, 10+j*7, 14, 14)
		}
	}
}

func (gol *GOL) ProcessCutText(conn *gorfb.RFBConn, text string) {
	fmt.Printf("Text send from client: %s\n", text)
}

func main() {
	fmt.Println("Conway's Game Of Life presented through a VNC (RFB) session at port 5901")
	gol := &GOL{}

	rfb := gorfb.RFBServer{
		PixelFormat: gorfb.PixelFormat{
			BitsPerPixel: 32,
			Depth:        32,
			BigEndian:    1,
			TrueColor:    1,
			RedMax:       255,
			GreenMax:     255,
			BlueMax:      255,
			RedShift:     24,
			GreenShift:   16,
			BlueShift:    8,
		},
		Port:         "5901",
		Width:        1366,
		Height:       768,
		Handler:      gol,
		Authenticate: true,
		AuthText:     "conway12", // This is the authentication string that must be entered on the client
	}
	err := rfb.StartServer()
	if err != nil {
		fmt.Println(err)
	}
}
