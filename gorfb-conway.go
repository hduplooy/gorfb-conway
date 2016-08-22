// gorfb-conway project main.go
package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"time"

	"github.com/hduplooy/gorfb"
)

type GOL struct { // Game Of Life data
	Img     *image.RGBA
	board   [][]byte
	running bool
	RFBConn *gorfb.RFBConn
}

var black = color.RGBA{0, 0, 0, 0}
var white = color.RGBA{255, 255, 255, 0}

func (gol *GOL) sendRectangle(x, y, width, height int) error {
	if x == 0 && y == 0 && width == 1366 && height == 768 {
		return gol.RFBConn.SendRectangle(0, 0, 1366, 768, gol.Img.Pix)
	}
	sz := width * height * 4
	buf := make([]byte, sz)
	if x == 0 && width == 1366 {
		copy(buf[0:], gol.Img.Pix[y*1366*4:(y+height-1)*1366*4])
	} else {
		st1 := (y*1366 + x) * 4
		st2 := 0
		w4 := width * 4
		for i := 0; i < height; i, st1, st2 = i+1, st1+1366*4, st2+w4 {
			copy(buf[st2:], gol.Img.Pix[st1:st1+w4])
		}
	}
	return gol.RFBConn.SendRectangle(x, y, width, height, buf)
}

func (gol *GOL) DrawHLine(x1, x2, y int, col color.RGBA) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if x2 < 0 || x1 >= 1366 || y < 0 || y >= 768 {
		return
	}
	if x1 < 0 {
		x1 = 0
	}
	if x2 >= 1366 {
		x2 = 1365
	}
	strt := (y*1366 + x1) * 4
	end := strt + (x2-x1)*4
	for strt <= end {
		gol.Img.Pix[strt] = col.R
		gol.Img.Pix[strt+1] = col.G
		gol.Img.Pix[strt+2] = col.B
		strt += 4
	}
}

func (gol *GOL) DrawVLine(x, y1, y2 int, col color.RGBA) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x < 0 || x >= 1365 || y2 < 0 || y1 >= 767 {
		return
	}
	if y1 < 0 {
		y1 = 0
	}
	if y2 >= 768 {
		y2 = 767
	}
	strt := (y1*1366 + x) * 4
	hght := y2 - y1
	for j := 0; j < hght; j++ {
		gol.Img.Pix[strt] = col.R
		gol.Img.Pix[strt+1] = col.G
		gol.Img.Pix[strt+2] = col.B
		strt += 1366 * 4
	}
}

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

func (gol *GOL) FillRect(x1, y1, x2, y2 int, col color.RGBA) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x2 < 0 || x1 >= 1366 || y2 < 0 || y1 >= 768 {
		return
	}
	if x1 < 0 {
		x1 = 0
	}
	if x2 >= 1366 {
		x2 = 1365
	}
	if y1 < 0 {
		y1 = 0
	}
	if y2 >= 768 {
		y2 = 767
	}
	strt := (y1*1366 + x1) * 4
	hght := y2 - y1
	wdth := x2 - x1
	for j := 0; j < hght; j++ {
		strt2 := strt
		for i := 0; i < wdth; i++ {
			gol.Img.Pix[strt2] = col.R
			gol.Img.Pix[strt2+1] = col.G
			gol.Img.Pix[strt2+2] = col.B
			strt2 += 4
		}
		strt += 1366 * 4
	}
}

func (gol *GOL) Update() {
	gol.FillRect(0, 0, 1366, 768, white)
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			if gol.board[i][j] == 1 {
				gol.FillRect(10+i*7, 10+j*7, 17+i*7, 17+j*7, black)
			}
		}
	}
	for i := 0; i <= 700; i += 7 {
		gol.DrawHLine(10, 710, i+10, black)
		gol.DrawVLine(i+10, 10, 710, black)
	}
}

func (gol *GOL) UpdateCell(i, j int) {
	if gol.board[i][j] == 1 {
		gol.FillRect(10+i*7, 10+j*7, 17+i*7, 17+j*7, black)
	} else {
		gol.FillRect(10+i*7, 10+j*7, 17+i*7, 17+j*7, white)
		gol.DrawRect(10+i*7, 10+j*7, 17+i*7, 17+j*7, black)
	}
}

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

func (gol *GOL) Run() {
	var oldboard [][]byte = make([][]byte, 100)
	for i := 0; i < 100; i++ {
		oldboard[i] = make([]byte, 100)
	}
	for {
		<-time.After(time.Nanosecond * 300000000) // .3 Seconds for every step
		// Make a copy of the board to work with the previous instance separate from the new one
		if gol.running {
			for i := 0; i < 100; i++ {
				copy(oldboard[i], gol.board[i])
			}
			for i := 0; i < 100; i++ {
				for j := 0; j < 100; j++ {
					cnt := checkCnt(oldboard, i, j)
					if oldboard[i][j] == 0 {
						if cnt == 3 {
							gol.board[i][j] = 1
						}
					} else {
						if cnt < 2 || cnt > 3 {
							gol.board[i][j] = 0
						}
					}
					if gol.board[i][j] != oldboard[i][j] {
						gol.UpdateCell(i, j)
					}
				}
			}
			if gol.sendRectangle(10, 10, 710, 710) != nil {
				break
			}
		}
	}

}

func (gol *GOL) Init(conn *gorfb.RFBConn) {
	gol.RFBConn = conn
	gol.Img = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1366, 768}})
	gol.board = make([][]byte, 100)
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

func (gol *GOL) ProcessSetPixelFormat(conn *gorfb.RFBConn, pf gorfb.PixelFormat) {
	// Ignored for now
}

func (gol *GOL) ProcessSetEncoding(conn *gorfb.RFBConn, encodings []int) {
	// Ignored for now
}

func (gol *GOL) ProcessUpdateRequest(conn *gorfb.RFBConn, x, y, width, height int, incremental bool) {
	if !incremental {
		gol.sendRectangle(x, y, width, height)
	}
}

func (gol *GOL) ProcessKeyEvent(conn *gorfb.RFBConn, key int, downflag bool) {
	if key == 'p' && downflag {
		gol.running = !gol.running
	}
}

func (gol *GOL) ProcessPointerEvent(conn *gorfb.RFBConn, x, y, button int) {
	if button >= 1 && x >= 10 && x < 710 && y >= 10 && y < 710 {
		i := (x - 10) / 7
		j := (y - 10) / 7
		if button == 1 {
			gol.board[i][j] = 1
		} else {
			gol.board[i][j] = 0
		}
		gol.UpdateCell(i, j)
		if !gol.running {
			gol.sendRectangle(10+i*7, 10+j*7, 14, 14)
		}
	}
}

func (gol *GOL) ProcessCutText(conn *gorfb.RFBConn, text string) {
	// Ignored for now
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
		Port:    "5901",
		Width:   1366,
		Height:  768,
		Handler: gol,
	}
	rfb.StartServer()
}
